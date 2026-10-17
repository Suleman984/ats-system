"use client";

import { useEffect, useState } from "react";
import { teamAPI, errorMessage, type AdminRole, type TeamInvite, type TeamMember } from "@/lib/api";
import { useAuthStore } from "@/lib/store";
import { useIsOwner } from "@/lib/dashboard-path";
import { toast } from "@/components/Toast";

export default function TeamPage() {
  const isOwner = useIsOwner();
  const currentId = useAuthStore((state) => state.user?.id);
  const [members, setMembers] = useState<TeamMember[]>([]);
  const [invites, setInvites] = useState<TeamInvite[]>([]);
  const [loading, setLoading] = useState(true);
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<AdminRole>("recruiter");
  const [sending, setSending] = useState(false);

  const load = () =>
    teamAPI
      .get()
      .then(({ data }) => {
        setMembers(data.members);
        setInvites(data.invites);
      })
      .catch((error) => toast.error(errorMessage(error, "Failed to load team")))
      .finally(() => setLoading(false));

  useEffect(() => {
    if (isOwner) load();
    else setLoading(false);
  }, [isOwner]);

  const handleInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    setSending(true);
    try {
      await teamAPI.invite(email, role);
      toast.success(`Invite sent to ${email}`);
      setEmail("");
      load();
    } catch (error) {
      toast.error(errorMessage(error, "Failed to send invite"));
    } finally {
      setSending(false);
    }
  };

  const act = async (action: () => Promise<unknown>, confirmText: string | null, success: string) => {
    if (confirmText && !window.confirm(confirmText)) return;
    try {
      await action();
      toast.success(success);
      load();
    } catch (error) {
      toast.error(errorMessage(error, "Something went wrong"));
    }
  };

  if (!isOwner) {
    return (
      <div className="p-6 text-gray-600">Only account owners can manage the team.</div>
    );
  }
  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-4xl mx-auto space-y-6">
        <div>
          <h1 className="text-3xl font-bold mb-1">Team</h1>
          <p className="text-gray-600">
            Owners manage the team, billing and deletions. Recruiters handle jobs and candidates.
          </p>
        </div>

        <form onSubmit={handleInvite} className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Invite a teammate</h2>
          <div className="flex flex-col sm:flex-row gap-3">
            <input
              type="email"
              required
              placeholder="name@company.com"
              className="flex-1 px-4 py-2 border rounded-lg"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
            <select
              className="px-4 py-2 border rounded-lg"
              value={role}
              onChange={(e) => setRole(e.target.value as AdminRole)}
            >
              <option value="recruiter">Recruiter</option>
              <option value="owner">Owner</option>
            </select>
            <button
              type="submit"
              disabled={sending}
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
            >
              {sending ? "Sending..." : "Send invite"}
            </button>
          </div>
        </form>

        <div className="bg-white rounded-lg shadow">
          <h2 className="text-lg font-semibold p-6 pb-2">Members</h2>
          <ul className="divide-y">
            {members.map((m) => (
              <li key={m.id} className="flex flex-wrap items-center justify-between gap-3 px-6 py-4">
                <div>
                  <p className="font-medium">
                    {m.name} {m.id === currentId && <span className="text-xs text-gray-500">(you)</span>}
                  </p>
                  <p className="text-sm text-gray-500">{m.email}</p>
                </div>
                <div className="flex items-center gap-3">
                  <select
                    className="px-2 py-1 border rounded text-sm"
                    value={m.role}
                    disabled={m.id === currentId}
                    onChange={(e) =>
                      act(
                        () => teamAPI.setRole(m.id, e.target.value as AdminRole),
                        null,
                        `${m.name} is now ${e.target.value}`
                      )
                    }
                  >
                    <option value="owner">Owner</option>
                    <option value="recruiter">Recruiter</option>
                  </select>
                  {m.id !== currentId && (
                    <button
                      onClick={() =>
                        act(
                          () => teamAPI.remove(m.id),
                          `Remove ${m.name}? They'll be signed out and lose access immediately.`,
                          `${m.name} removed`
                        )
                      }
                      className="text-sm text-red-600 hover:underline"
                    >
                      Remove
                    </button>
                  )}
                </div>
              </li>
            ))}
          </ul>
        </div>

        {invites.length > 0 && (
          <div className="bg-white rounded-lg shadow">
            <h2 className="text-lg font-semibold p-6 pb-2">Pending invites</h2>
            <ul className="divide-y">
              {invites.map((inv) => (
                <li key={inv.id} className="flex flex-wrap items-center justify-between gap-3 px-6 py-4">
                  <div>
                    <p className="font-medium">{inv.email}</p>
                    <p className="text-sm text-gray-500">
                      {inv.role} · expires {new Date(inv.expires_at).toLocaleDateString()}
                    </p>
                  </div>
                  <div className="flex gap-3 text-sm">
                    <button
                      onClick={() => act(() => teamAPI.invite(inv.email, inv.role), null, "Invite re-sent")}
                      className="text-blue-600 hover:underline"
                    >
                      Resend
                    </button>
                    <button
                      onClick={() => act(() => teamAPI.revokeInvite(inv.id), `Revoke the invite for ${inv.email}?`, "Invite revoked")}
                      className="text-red-600 hover:underline"
                    >
                      Revoke
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </div>
  );
}
