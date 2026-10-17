"use client";

import { useState } from "react";
import { crmAPI, errorMessage, type CandidateNote } from "@/lib/api";
import { toast } from "@/components/Toast";
import { useAuthStore } from "@/lib/store";

export default function NotesModal({
  applicationId,
  notes,
  onClose,
  onRefresh,
}: {
  applicationId: string;
  notes: CandidateNote[];
  onClose: () => void;
  onRefresh: () => void;
}) {
  const [newNote, setNewNote] = useState("");
  const [isPrivate, setIsPrivate] = useState(false);
  const [loading, setLoading] = useState(false);
  const currentAdminId = useAuthStore((state) => state.user?.id);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editText, setEditText] = useState("");
  const [editPrivate, setEditPrivate] = useState(false);

  const startEdit = (note: CandidateNote) => {
    setEditingId(note.id);
    setEditText(note.note);
    setEditPrivate(note.is_private);
  };

  const handleSaveEdit = async () => {
    if (!editingId || !editText.trim()) return;
    setLoading(true);
    try {
      await crmAPI.updateNote(editingId, editText.trim(), editPrivate);
      toast.success("Note updated");
      setEditingId(null);
      onRefresh();
    } catch (error) {
      toast.error(errorMessage(error, "Failed to update note"));
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (noteId: string) => {
    if (!window.confirm("Delete this note? This can't be undone.")) return;
    try {
      await crmAPI.deleteNote(noteId);
      toast.success("Note deleted");
      onRefresh();
    } catch (error) {
      toast.error(errorMessage(error, "Failed to delete note"));
    }
  };

  const handleAddNote = async () => {
    if (!newNote.trim()) {
      toast.error("Please enter a note");
      return;
    }
    setLoading(true);
    try {
      await crmAPI.addNote(applicationId, newNote.trim(), isPrivate);
      toast.success("Note added successfully");
      setNewNote("");
      setIsPrivate(false);
      onRefresh();
    } catch (error) {
      toast.error(errorMessage(error, "Failed to add note"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <div className="p-6 border-b">
          <div className="flex justify-between items-center">
            <h2 className="text-2xl font-bold">Candidate Notes</h2>
            <button
              onClick={onClose}
              className="text-gray-500 hover:text-gray-700 text-2xl"
            >
              ×
            </button>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto p-6">
          <div className="space-y-4 mb-6">
            {notes.length === 0 ? (
              <p className="text-gray-500 text-center py-8">No notes yet</p>
            ) : (
              notes.map((note) => (
                <div key={note.id} className="border rounded-lg p-4 bg-gray-50">
                  <div className="flex justify-between items-start mb-2">
                    <div>
                      <p className="font-semibold text-sm">
                        {note.admin?.name || "Unknown"}
                      </p>
                      <p className="text-xs text-gray-500">
                        {new Date(note.created_at).toLocaleString()}
                        {note.is_private && (
                          <span className="ml-2 text-blue-600">🔒 Private</span>
                        )}
                      </p>
                    </div>
                    {/* Only the author can change a note; the API enforces this too. */}
                    {note.admin_id === currentAdminId && editingId !== note.id && (
                      <div className="flex gap-3 text-xs">
                        <button
                          onClick={() => startEdit(note)}
                          className="text-blue-600 hover:underline"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleDelete(note.id)}
                          className="text-red-600 hover:underline"
                        >
                          Delete
                        </button>
                      </div>
                    )}
                  </div>
                  {editingId === note.id ? (
                    <div>
                      <textarea
                        className="w-full px-3 py-2 border rounded-lg resize-none mb-2 bg-white"
                        rows={3}
                        value={editText}
                        onChange={(e) => setEditText(e.target.value)}
                      />
                      <div className="flex items-center justify-between">
                        <label className="flex items-center text-sm text-gray-600">
                          <input
                            type="checkbox"
                            checked={editPrivate}
                            onChange={(e) => setEditPrivate(e.target.checked)}
                            className="mr-2"
                          />
                          Private
                        </label>
                        <div className="flex gap-2">
                          <button
                            onClick={() => setEditingId(null)}
                            className="px-3 py-1 text-sm text-gray-600 hover:text-gray-800"
                          >
                            Cancel
                          </button>
                          <button
                            onClick={handleSaveEdit}
                            disabled={loading || !editText.trim()}
                            className="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-blue-300"
                          >
                            Save
                          </button>
                        </div>
                      </div>
                    </div>
                  ) : (
                    <p className="text-gray-800 whitespace-pre-wrap">
                      {note.note}
                    </p>
                  )}
                </div>
              ))
            )}
          </div>

          <div className="border-t pt-4">
            <h3 className="font-semibold mb-3">Add New Note</h3>
            <textarea
              className="w-full px-4 py-2 border rounded-lg resize-none mb-3"
              placeholder="Enter your note here..."
              rows={4}
              value={newNote}
              onChange={(e) => setNewNote(e.target.value)}
            />
            <div className="flex items-center justify-between">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={isPrivate}
                  onChange={(e) => setIsPrivate(e.target.checked)}
                  className="mr-2"
                />
                <span className="text-sm text-gray-600">
                  Private note (only visible to you)
                </span>
              </label>
              <button
                onClick={handleAddNote}
                disabled={loading || !newNote.trim()}
                className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-300"
              >
                {loading ? "Adding..." : "Add Note"}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
