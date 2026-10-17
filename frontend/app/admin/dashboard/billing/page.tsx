"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  billingAPI,
  errorMessage,
  type Payment,
  type PlanUsage,
  type SubscriptionPlan,
} from "@/lib/api";
import { useIsOwner } from "@/lib/dashboard-path";
import { toast } from "@/components/Toast";

type BillingData = Awaited<ReturnType<typeof billingAPI.get>>["data"];

const UNLIMITED = 100000;
const limitText = (n: number) => (n >= UNLIMITED ? "Unlimited" : n.toLocaleString());

const parseFeatures = (features: SubscriptionPlan["features"]): string[] => {
  if (Array.isArray(features)) return features;
  try {
    return JSON.parse(features);
  } catch {
    return [];
  }
};

function UsageBar({ label, used, limit }: { label: string; used: number; limit: number }) {
  const pct = limit >= UNLIMITED ? 0 : Math.min(100, Math.round((used / Math.max(limit, 1)) * 100));
  return (
    <div>
      <div className="flex justify-between text-sm mb-1">
        <span>{label}</span>
        <span className="text-gray-600">
          {used.toLocaleString()} / {limitText(limit)}
        </span>
      </div>
      <div className="h-2 bg-gray-100 rounded">
        <div
          className={`h-2 rounded ${pct >= 100 ? "bg-red-500" : pct >= 80 ? "bg-yellow-500" : "bg-blue-600"}`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}

function Billing() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const isOwner = useIsOwner();
  const [data, setData] = useState<BillingData | null>(null);
  const [loading, setLoading] = useState(true);
  const [buying, setBuying] = useState<string | null>(null);

  const load = useCallback(
    () =>
      billingAPI
        .get()
        .then(({ data }) => setData(data))
        .catch((error) => toast.error(errorMessage(error, "Failed to load billing"))),
    []
  );

  // Returning from Safepay: confirm with the server (it re-checks with Safepay) before reloading.
  useEffect(() => {
    const paymentId = searchParams.get("payment");
    const cancelled = searchParams.get("cancelled");
    const run = async () => {
      if (paymentId && cancelled) {
        toast.info("Checkout cancelled. You haven't been charged.");
      } else if (paymentId) {
        try {
          const { data } = await billingAPI.confirm(paymentId);
          if (data.payment.status === "completed") toast.success("Payment received. Your plan is active.");
          else if (data.payment.status === "failed") toast.error("The payment didn't go through.");
          else toast.info("Payment is processing. This page will show your new plan once it's confirmed.");
        } catch (error) {
          toast.error(errorMessage(error, "Couldn't confirm the payment"));
        }
      }
      if (paymentId) router.replace("/admin/dashboard/billing");
      await load();
      setLoading(false);
    };
    run();
  }, [searchParams, router, load]);

  const handleBuy = async (plan: SubscriptionPlan) => {
    setBuying(plan.name);
    try {
      const { data } = await billingAPI.checkout(plan.name);
      window.location.assign(data.checkout_url);
    } catch (error) {
      toast.error(errorMessage(error, "Couldn't start checkout"));
      setBuying(null);
    }
  };

  if (loading || !data) return <div className="p-6">Loading...</div>;
  const { current_plan: current, usage, subscription } = data;
  const u: PlanUsage = usage;

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-5xl mx-auto space-y-6">
        <h1 className="text-3xl font-bold">Billing</h1>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex flex-wrap justify-between gap-2 mb-4">
            <div>
              <p className="text-sm text-gray-500">Current plan</p>
              <p className="text-2xl font-bold">{current.name}</p>
            </div>
            {subscription && (
              <p className="text-sm text-gray-600 self-end">
                Paid through {new Date(subscription.current_period_end).toLocaleDateString()}
              </p>
            )}
          </div>
          <div className="grid sm:grid-cols-3 gap-6">
            <UsageBar label="Open jobs" used={u.open_jobs} limit={current.max_jobs} />
            <UsageBar label="Applications this month" used={u.applications_this_month} limit={current.max_applications} />
            <UsageBar label="Team seats (incl. invites)" used={u.admins + u.pending_invites} limit={current.max_admins} />
          </div>
          {!current.ai_shortlisting && (
            <p className="text-sm text-gray-600 mt-4">CV scoring isn&apos;t included in this plan.</p>
          )}
        </div>

        <div className="grid md:grid-cols-3 gap-6">
          {data.plans.map((plan) => {
            const isCurrent = plan.id === current.id;
            return (
              <div
                key={plan.id}
                className={`bg-white rounded-lg shadow p-6 flex flex-col ${isCurrent ? "ring-2 ring-blue-600" : ""}`}
              >
                <h2 className="text-xl font-bold">{plan.name}</h2>
                <p className="text-3xl font-bold my-2">
                  PKR {plan.price_pkr.toLocaleString()}
                  <span className="text-sm font-normal text-gray-500"> / month</span>
                </p>
                <ul className="text-sm text-gray-700 space-y-1 mb-6 flex-1">
                  {parseFeatures(plan.features).map((f) => (
                    <li key={f}>✓ {f}</li>
                  ))}
                </ul>
                {isOwner ? (
                  <button
                    onClick={() => handleBuy(plan)}
                    disabled={!data.checkout_available || buying !== null || plan.price_pkr <= 0}
                    className="w-full py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700 disabled:bg-blue-300"
                  >
                    {buying === plan.name
                      ? "Redirecting..."
                      : isCurrent
                        ? "Renew 1 month"
                        : `Switch to ${plan.name}`}
                  </button>
                ) : (
                  isCurrent && <p className="text-center text-sm text-gray-500">Your current plan</p>
                )}
              </div>
            );
          })}
        </div>
        {!data.checkout_available && (
          <p className="text-sm text-gray-500">Online payments aren&apos;t configured yet.</p>
        )}
        {!isOwner && <p className="text-sm text-gray-500">Only account owners can change the plan.</p>}

        {data.payments.length > 0 && (
          <div className="bg-white rounded-lg shadow">
            <h2 className="text-lg font-semibold p-6 pb-2">Payment history</h2>
            <table className="w-full text-sm">
              <tbody className="divide-y">
                {data.payments.map((p: Payment) => (
                  <tr key={p.id}>
                    <td className="px-6 py-3">{new Date(p.created_at).toLocaleDateString()}</td>
                    <td className="px-6 py-3">
                      {p.currency} {p.amount.toLocaleString()}
                    </td>
                    <td className="px-6 py-3 capitalize">{p.status}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

// useSearchParams needs a Suspense boundary so the page can prerender.
export default function BillingPage() {
  return (
    <Suspense>
      <Billing />
    </Suspense>
  );
}
