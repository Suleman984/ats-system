"use client";

import { useState } from "react";
import { crmAPI, errorMessage, type Application } from "@/lib/api";
import { toast } from "@/components/Toast";

export default function ReferralModal({
  application,
  onClose,
  onSave,
}: {
  application: Application | undefined;
  onClose: () => void;
  onSave: () => void;
}) {
  const [referralSource, setReferralSource] = useState(
    application?.referral_source || ""
  );
  const [referredByName, setReferredByName] = useState(
    application?.referred_by_name || ""
  );
  const [referredByEmail, setReferredByEmail] = useState(
    application?.referred_by_email || ""
  );
  const [referredByPhone, setReferredByPhone] = useState(
    application?.referred_by_phone || ""
  );
  const [loading, setLoading] = useState(false);

  const handleSave = async () => {
    if (!application) return;
    setLoading(true);
    try {
      await crmAPI.updateReferral(
        application.id,
        referralSource,
        referredByName,
        referredByEmail,
        referredByPhone
      );
      toast.success("Referral information updated");
      onSave();
    } catch (error) {
      toast.error(
        errorMessage(error, "Failed to update referral info")
      );
    } finally {
      setLoading(false);
    }
  };

  if (!application) return null;

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
        <div className="p-6 border-b">
          <div className="flex justify-between items-center">
            <h2 className="text-2xl font-bold">Referral Information</h2>
            <button
              onClick={onClose}
              className="text-gray-500 hover:text-gray-700 text-2xl"
            >
              ×
            </button>
          </div>
        </div>
        <div className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">
              Referral Source
            </label>
            <input
              type="text"
              className="w-full px-4 py-2 border rounded-lg"
              placeholder="e.g., LinkedIn, Employee Referral, Job Board"
              value={referralSource}
              onChange={(e) => setReferralSource(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">
              Referred By (Name)
            </label>
            <input
              type="text"
              className="w-full px-4 py-2 border rounded-lg"
              placeholder="Name of referrer"
              value={referredByName}
              onChange={(e) => setReferredByName(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">
              Referred By (Email)
            </label>
            <input
              type="email"
              className="w-full px-4 py-2 border rounded-lg"
              placeholder="email@example.com"
              value={referredByEmail}
              onChange={(e) => setReferredByEmail(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">
              Referred By (Phone)
            </label>
            <input
              type="tel"
              className="w-full px-4 py-2 border rounded-lg"
              placeholder="+1234567890"
              value={referredByPhone}
              onChange={(e) => setReferredByPhone(e.target.value)}
            />
          </div>
          <div className="flex gap-3 pt-4">
            <button
              onClick={onClose}
              className="flex-1 px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              disabled={loading}
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
            >
              {loading ? "Saving..." : "Save"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
