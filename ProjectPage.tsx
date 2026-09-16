import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import { Link, useParams } from "react-router-dom";
import { api, ApiError } from "../lib/api";
import type { Keyword, Project, Report } from "../lib/types";
import {
  Button,
  Card,
  ErrorText,
  Input,
  StatusBadge,
} from "../components/ui";

const MAX_KEYWORDS = 5;

export function ProjectPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [project, setProject] = useState<Project | null>(null);
  const [keywords, setKeywords] = useState<Keyword[] | null>(null);
  const [reports, setReports] = useState<Report[] | null>(null);
  const [newKeyword, setNewKeyword] = useState("");
  const [keywordError, setKeywordError] = useState("");
  const [addingKeyword, setAddingKeyword] = useState(false);
  const [requestingReport, setRequestingReport] = useState(false);
  const [reportError, setReportError] = useState("");

  function refreshAll() {
    if (!projectId) return;
    api.getProject(projectId).then(setProject);
    api.listKeywords(projectId).then(setKeywords);
    api.listReports(projectId).then(setReports);
  }

  useEffect(refreshAll, [projectId]);

  // Poll while any report is pending/running so status updates automatically.
  useEffect(() => {
    if (!reports || !projectId) return;
    const hasActive = reports.some(
      (r) => r.status === "pending" || r.status === "running",
    );
    if (!hasActive) return;
    const id = setInterval(() => {
      api.listReports(projectId).then(setReports);
    }, 4000);
    return () => clearInterval(id);
  }, [reports, projectId]);

  async function handleAddKeyword(e: FormEvent) {
    e.preventDefault();
    if (!projectId) return;
    setKeywordError("");
    setAddingKeyword(true);
    try {
      await api.addKeyword(projectId, newKeyword);
      setNewKeyword("");
      api.listKeywords(projectId).then(setKeywords);
    } catch (err) {
      setKeywordError(
        err instanceof ApiError ? err.message : "Failed to add keyword",
      );
    } finally {
      setAddingKeyword(false);
    }
  }

  async function handleDeleteKeyword(keywordId: string) {
    if (!projectId) return;
    await api.deleteKeyword(projectId, keywordId);
    api.listKeywords(projectId).then(setKeywords);
  }

  async function handleRequestReport() {
    if (!projectId) return;
    setReportError("");
    setRequestingReport(true);
    try {
      await api.requestReport(projectId);
      api.listReports(projectId).then(setReports);
    } catch (err) {
      setReportError(
        err instanceof ApiError ? err.message : "Failed to request report",
      );
    } finally {
      setRequestingReport(false);
    }
  }

  if (!project || !keywords || !reports) {
    return <p className="text-sm text-neutral-400">Loading...</p>;
  }

  const canAddKeyword = keywords.length < MAX_KEYWORDS;
  const canRunReport = keywords.length > 0;

  return (
    <div className="space-y-8">
      <div>
        <Link to="/" className="text-sm text-neutral-500 hover:underline">
          ← All projects
        </Link>
        <h1 className="mt-2 text-2xl font-semibold">{project.name}</h1>
        <p className="mt-1 max-w-2xl text-neutral-600">
          {project.product_description}
        </p>
        {project.target_url && (
          <a
            href={project.target_url}
            target="_blank"
            rel="noreferrer"
            className="mt-1 inline-block text-sm text-orange-600 hover:underline"
          >
            {project.target_url}
          </a>
        )}
      </div>

      <Card>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="font-semibold">
            Keywords ({keywords.length}/{MAX_KEYWORDS})
          </h2>
        </div>

        <div className="mb-4 flex flex-wrap gap-2">
          {keywords.map((k) => (
            <span
              key={k.id}
              className="flex items-center gap-2 rounded-full bg-neutral-100 px-3 py-1 text-sm"
            >
              {k.keyword}
              <button
                onClick={() => handleDeleteKeyword(k.id)}
                className="text-neutral-400 hover:text-red-600"
                aria-label={`Remove ${k.keyword}`}
              >
                ×
              </button>
            </span>
          ))}
          {keywords.length === 0 && (
            <p className="text-sm text-neutral-400">No keywords yet.</p>
          )}
        </div>

        {canAddKeyword && (
          <form onSubmit={handleAddKeyword} className="flex gap-2">
            <Input
              placeholder="e.g. notion alternative"
              value={newKeyword}
              onChange={(e) => setNewKeyword(e.target.value)}
            />
            <Button type="submit" disabled={addingKeyword}>
              Add
            </Button>
          </form>
        )}
        <ErrorText>{keywordError}</ErrorText>
      </Card>

      <Card>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="font-semibold">Reports</h2>
          <Button
            onClick={handleRequestReport}
            disabled={!canRunReport || requestingReport}
          >
            {requestingReport ? "Requesting..." : "Find new leads"}
          </Button>
        </div>
        {!canRunReport && (
          <p className="mb-4 text-sm text-neutral-400">
            Add at least one keyword before running a search.
          </p>
        )}
        <ErrorText>{reportError}</ErrorText>

        {reports.length === 0 && (
          <p className="text-sm text-neutral-400">No reports yet.</p>
        )}

        <ul className="divide-y divide-neutral-100">
          {reports.map((r) => (
            <li key={r.id} className="flex items-center justify-between py-3">
              <div>
                <p className="text-sm text-neutral-700">
                  {new Date(r.created_at).toLocaleString()}
                </p>
                {r.status === "failed" && r.error && (
                  <p className="text-xs text-red-500">{r.error}</p>
                )}
              </div>
              <div className="flex items-center gap-3">
                <StatusBadge status={r.status} />
                <Link
                  to={`/reports/${r.id}`}
                  className="text-sm font-medium text-orange-600 hover:underline"
                >
                  View
                </Link>
              </div>
            </li>
          ))}
        </ul>
      </Card>
    </div>
  );
}
