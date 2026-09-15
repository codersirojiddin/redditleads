import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api } from "../lib/api";
import type { ReportDetail } from "../lib/types";
import { Card, Spinner, StatusBadge } from "../components/ui";

function ScorePill({ label, value }: { label: string; value: number }) {
  return (
    <span className="rounded-md bg-neutral-100 px-2 py-1 text-xs text-neutral-600">
      {label}: <span className="font-medium text-neutral-800">{Math.round(value)}</span>
    </span>
  );
}

export function ReportPage() {
  const { reportId } = useParams<{ reportId: string }>();
  const [report, setReport] = useState<ReportDetail | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  function refresh() {
    if (!reportId) return;
    api.getReport(reportId).then(setReport);
  }

  useEffect(refresh, [reportId]);

  useEffect(() => {
    if (!report) return;
    if (report.status !== "pending" && report.status !== "running") return;
    const id = setInterval(refresh, 4000);
    return () => clearInterval(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [report?.status, reportId]);

  async function copyReply(oppId: string, text: string) {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedId(oppId);
      setTimeout(() => setCopiedId(null), 1500);
    } catch {
      // clipboard API might be unavailable; silently ignore
    }
  }

  if (!report) {
    return <p className="text-sm text-neutral-400">Loading...</p>;
  }

  return (
    <div className="space-y-6">
      <div>
        <Link
          to={`/projects/${report.project_id}`}
          className="text-sm text-neutral-500 hover:underline"
        >
          ← Back to project
        </Link>
        <div className="mt-2 flex items-center gap-3">
          <h1 className="text-2xl font-semibold">Report</h1>
          <StatusBadge status={report.status} />
        </div>
        <p className="text-sm text-neutral-400">
          Requested {new Date(report.created_at).toLocaleString()}
        </p>
      </div>

      {(report.status === "pending" || report.status === "running") && (
        <Card className="flex items-center gap-3 text-neutral-600">
          <Spinner className="h-5 w-5" />
          <p>
            Searching Reddit and scoring posts with AI — this can take a
            minute. This page refreshes automatically.
          </p>
        </Card>
      )}

      {report.status === "failed" && (
        <Card className="border-red-200 bg-red-50 text-red-700">
          <p className="font-medium">This report failed.</p>
          {report.error && <p className="mt-1 text-sm">{report.error}</p>}
        </Card>
      )}

      {report.status === "completed" && report.opportunities.length === 0 && (
        <Card className="text-center text-neutral-500">
          No qualifying opportunities found for this run. Try broadening your
          keywords.
        </Card>
      )}

      <div className="space-y-4">
        {report.opportunities.map((o, idx) => (
          <Card key={o.id}>
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0">
                <p className="text-xs text-neutral-400">
                  #{idx + 1} · r/{o.subreddit} · u/{o.author}
                </p>
                <a
                  href={o.post_url}
                  target="_blank"
                  rel="noreferrer"
                  className="font-medium text-neutral-900 hover:text-orange-600 hover:underline"
                >
                  {o.post_title}
                </a>
              </div>
              <div className="flex shrink-0 flex-col items-end gap-1">
                <span className="text-lg font-semibold text-orange-600">
                  {Math.round(o.total_score)}
                </span>
                <div className="flex gap-1">
                  <ScorePill label="Relevance" value={o.relevance_score} />
                  <ScorePill label="Intent" value={o.intent_score} />
                </div>
              </div>
            </div>

            <p className="mt-3 text-sm text-neutral-600">{o.reasoning}</p>

            {o.suggested_reply && (
              <div className="mt-4 rounded-lg bg-neutral-50 p-3">
                <div className="mb-1 flex items-center justify-between">
                  <p className="text-xs font-medium text-neutral-500">
                    Suggested reply
                  </p>
                  <button
                    onClick={() => copyReply(o.id, o.suggested_reply)}
                    className="text-xs font-medium text-orange-600 hover:underline"
                  >
                    {copiedId === o.id ? "Copied!" : "Copy"}
                  </button>
                </div>
                <p className="text-sm text-neutral-700">
                  {o.suggested_reply}
                </p>
              </div>
            )}
          </Card>
        ))}
      </div>
    </div>
  );
}
