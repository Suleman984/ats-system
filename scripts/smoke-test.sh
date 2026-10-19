#!/usr/bin/env bash
# End-to-end smoke test against a running API backed by local Supabase with supabase/seed.sql loaded.
#   API_URL     default http://127.0.0.1:8080/api
#   DB_URL      default the local Supabase database
#   STORAGE_URL default http://127.0.0.1:54321 (to prove CV buckets are private)
#   SMOKE_CLAMAV=1 also checks that an infected upload is blocked (needs CLAMAV_ADDR on the API)
# Mutates data: run it against a local or CI database only, never production.
set -uo pipefail

A=${API_URL:-http://127.0.0.1:8080/api}
DB=${DB_URL:-postgresql://postgres:postgres@127.0.0.1:54322/postgres}
STORAGE=${STORAGE_URL:-http://127.0.0.1:54321}
JOB=33333333-3333-4333-8333-333333333331 # seeded "Senior Backend Engineer"

WORK=$(mktemp -d); trap 'kill $(jobs -p) 2>/dev/null; rm -rf "$WORK"' EXIT
cd "$WORK"
J=cookies.txt; FAILS=0; RUN=$RANDOM
pass(){ echo "PASS  $1"; }
fail(){ echo "FAIL  $1 -> $2"; FAILS=$((FAILS+1)); }
# Exactly three arguments, or a mangled command line could let a check pass by accident.
check(){
  [ $# -eq 3 ] || { fail "$1" "check got $# arguments (quoting bug in the test)"; return; }
  [ "$2" = "$3" ] && pass "$1 ($2)" || fail "$1" "got '$2' want '$3'"
}
# JSON bodies are built in variables: macOS bash 3.2 brace-expands "{a,b}" inside "$(...)".
code(){ curl -s -o /dev/null -w '%{http_code}' "$@"; }
json(){ python3 -c "import sys,json;d=json.load(sys.stdin);print($1)"; }
sql(){ PGOPTIONS='-c client_min_messages=error' psql "$DB" -At -c "$1"; }
H=(-H Content-Type:application/json -H X-ATS-Client:web)

echo "== auth"
check "login demo admin" "$(code -c $J "${H[@]}" -d '{"email":"admin@acme.test","password":"demo12345"}' $A/auth/login)" 200
grep -q ats_session $J && pass "session cookie set" || fail "session cookie" missing
check "role from /auth/me" "$(curl -s -b $J $A/auth/me | json 'd["admin"]["role"]')" owner
check "write without client header" "$(code -b $J -X POST -H Content-Type:application/json -d '{}' $A/jobs)" 403
check "dashboard stats" "$(code -b $J $A/dashboard/stats)" 200

echo "== uploads"
printf '%%PDF-1.4\n1 0 obj << >> stream\nBT (Go engineer with PostgreSQL Docker Kubernetes and seven years building APIs. English fluent.) Tj ET\nendstream endobj\n%%%%EOF\n' > cv.pdf
printf 'hello' > fake.pdf
check "fake pdf rejected" "$(code -F file=@fake.pdf $A/upload/cv)" 400
check "real pdf accepted" "$(curl -s -o up.json -w '%{http_code}' -F file=@cv.pdf $A/upload/cv)" 201
UP=$(json 'd["upload_id"]' < up.json)
grep -q '"url"' up.json && fail "upload hides storage URL" leaked || pass "upload returns an ID, not a URL"

echo "== apply"
EMAIL="smoke$RUN@example.com"
body(){ echo "{\"job_id\":\"$JOB\",\"full_name\":\"Smoke Test\",\"email\":\"$1\",\"years_of_experience\":3,\"resume_upload_id\":\"$2\"}"; }
check "submit with upload" "$(curl -s -o app.json -w '%{http_code}' -H Content-Type:application/json -d "$(body $EMAIL $UP)" $A/applications)" 201
APP=$(json 'd["application"]["id"]' < app.json)
check "upload can't be reused" "$(code -H Content-Type:application/json -d "$(body other$RUN@example.com $UP)" $A/applications)" 400
UP2=$(curl -s -F file=@cv.pdf $A/upload/cv | json 'd["upload_id"]')
check "duplicate application" "$(code -H Content-Type:application/json -d "$(body $(echo "$EMAIL" | tr a-z A-Z) $UP2)" $A/applications)" 409

echo "== private files"
URL=$(curl -s -b $J $A/applications/$APP/files/resume | json 'd["url"]')
check "signed link serves the CV" "$(curl -s "$URL" | head -c 4)" "%PDF"
OBJ=$(sql "select resume_path from applications where id='$APP'")
check "public bucket URL blocked" "$(code $STORAGE/storage/v1/object/public/resumes/$OBJ)" 400
check "file link needs a session" "$(code $A/applications/$APP/files/resume)" 401

echo "== pipeline"
for step in shortlist advance advance advance; do code -b $J "${H[@]}" -X PUT $A/applications/$APP/$step >/dev/null; done
check "reaches hired" "$(sql "select status from applications where id='$APP'")" hired
check "can't advance past hired" "$(code -b $J "${H[@]}" -X PUT $A/applications/$APP/advance)" 409
check "reopen" "$(code -b $J "${H[@]}" -X PUT $A/applications/$APP/reopen)" 200

if [ "${SMOKE_CLAMAV:-0}" = 1 ]; then
  echo "== virus scanning"
  # The EICAR string is the industry-standard harmless test file every scanner flags.
  python3 - <<'PY'
import zipfile
eicar = 'X5O!P%@AP[4\\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*'
with zipfile.ZipFile('eicar.zip', 'w') as z:
    z.writestr('eicar.com', eicar)
PY
  check "infected portfolio blocked" "$(code -F file=@eicar.zip $A/upload/portfolio)" 422
fi

echo "== messaging, unread and live events"
unread(){ curl -s -b $J $A/messages/unread-summary | json 'd["by_job"].get("'$JOB'",0)'; }
BEFORE=$(unread)
TOKEN=$(curl -s -H Content-Type:application/json -d "{\"email\":\"$EMAIL\",\"application_id\":\"$APP\"}" $A/candidate/stream-token | json 'd["token"]')
curl -s -N -b $J $A/events > rec.txt &
curl -s -N "$A/candidate/events?token=$TOKEN" > cand.txt &
sleep 1
code -H Content-Type:application/json -d "{\"application_id\":\"$APP\",\"sender_email\":\"$EMAIL\",\"message\":\"hello\"}" $A/candidate/messages/send >/dev/null
sleep 1
check "unread summary counts it" "$(unread)" $((BEFORE+1))
check "unread filter finds it" "$(curl -s -b $J "$A/applications?unread=true&limit=50" | json 'any(a["id"]=="'$APP'" for a in d["applications"])')" True
code -b $J $A/applications/$APP/messages >/dev/null
code -b $J "${H[@]}" -d '{"message":"reply"}' $A/applications/$APP/messages >/dev/null
sleep 1
check "reading clears unread" "$(unread)" $BEFORE
check "recruiter stream got messages" "$(grep -c '^event: message' rec.txt)" 2
check "candidate stream got read receipt" "$(grep -c '^event: read' cand.txt)" 1
check "candidate never sees recruiter email" "$(grep -c 'admin@acme.test' cand.txt)" 0
check "bad stream token" "$(code "$A/candidate/events?token=nope")" 401

echo "== plans and billing (seed company is on Starter)"
check "CV scoring blocked" "$(code -b $J "${H[@]}" -d "{\"application_id\":\"$APP\"}" $A/applications/ai-shortlist)" 402
check "billing summary" "$(code -b $J $A/billing)" 200

echo "== signup and recovery"
SIGNUP=$(printf '{"company_name":"Smoke %s","email":"owner%s@smoke.test","password":"password123","name":"Owner"}' $RUN $RUN)
LOGIN=$(printf '{"email":"owner%s@smoke.test","password":"password123"}' $RUN)
check "register new" "$(code -H Content-Type:application/json -d "$SIGNUP" $A/auth/register)" 202
check "register existing (same answer)" "$(code -H Content-Type:application/json -d '{"company_name":"Dup","email":"admin@acme.test","password":"password123","name":"X"}' $A/auth/register)" 202
check "unverified login blocked" "$(curl -s -H Content-Type:application/json -d "$LOGIN" $A/auth/login | json 'd.get("code")')" email_unverified
check "forgot password" "$(code -H Content-Type:application/json -d '{"email":"admin@acme.test"}' $A/auth/forgot-password)" 200

echo "== public pages and cleanup"
check "sitemap entries" "$(curl -s $A/public/sitemap | json 'any(j["job_id"]=="'$JOB'" for j in d["jobs"])')" True
check "delete application" "$(code -b $J "${H[@]}" -X DELETE $A/applications/$APP)" 200
sleep 1
check "CV removed from storage" "$(sql "select count(*) from storage.objects where bucket_id='resumes' and name='$OBJ'")" 0
check "logout" "$(code -b $J -c $J "${H[@]}" -X POST $A/auth/logout)" 200
check "session gone after logout" "$(code -b $J $A/auth/me)" 401

echo
[ $FAILS -eq 0 ] && echo "All smoke checks passed." || echo "$FAILS smoke check(s) failed."
exit $FAILS
