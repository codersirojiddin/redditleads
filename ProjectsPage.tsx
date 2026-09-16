import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import { Link } from "react-router-dom";
import { api, ApiError } from "../lib/api";
import type { Project } from "../lib/types";
import {
  Button,
  Card,
  ErrorText,
  Input,
  Label,
  Textarea,
} from "../components/ui";

export function ProjectsPage() {
  const [projects, setProjects] = useState<Project[] | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [targetUrl, setTargetUrl] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  function refresh() {
    api.listProjects().then(setProjects).catch(() => setProjects([]));
  }

  useEffect(refresh, []);

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      await api.createProject({
        name,
        product_description: description,
        target_url: targetUrl || undefined,
      });
      setName("");
      setDescription("");
      setTargetUrl("");
      setShowForm(false);
      refresh();
    } catch (err) {
      setError(
        err instanceof ApiError ? err.message : "Failed to create project",
      );
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Your projects</h1>
        <Button onClick={() => setShowForm((v) => !v)}>
          {showForm ? "Cancel" : "New project"}
        </Button>
      </div>

      {showForm && (
        <Card className="mb-8">
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <Label>Project name</Label>
              <Input
                required
                placeholder="e.g. Notely"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>
            <div>
              <Label>Product description</Label>
              <Textarea
                required
                rows={3}
                placeholder="A lightweight note-taking app for people who find Notion too heavy..."
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
              <p className="mt-1 text-xs text-neutral-400">
                Used by the AI to judge whether a Reddit post is a real lead
                for this product.
              </p>
            </div>
            <div>
              <Label>Website (optional)</Label>
              <Input
                type="url"
                placeholder="https://yourproduct.com"
                value={targetUrl}
                onChange={(e) => setTargetUrl(e.target.value)}
              />
            </div>
            <ErrorText>{error}</ErrorText>
            <Button type="submit" disabled={submitting}>
              {submitting ? "Creating..." : "Create project"}
            </Button>
          </form>
        </Card>
      )}

      {projects === null && (
        <p className="text-sm text-neutral-400">Loading...</p>
      )}

      {projects && projects.length === 0 && !showForm && (
        <Card className="text-center text-neutral-500">
          <p>No projects yet.</p>
          <p className="text-sm">
            Create one to start finding leads on Reddit.
          </p>
        </Card>
      )}

      <div className="grid gap-4 sm:grid-cols-2">
        {projects?.map((p) => (
          <Link key={p.id} to={`/projects/${p.id}`}>
            <Card className="h-full transition hover:border-orange-300 hover:shadow-md">
              <h2 className="font-semibold">{p.name}</h2>
              <p className="mt-1 line-clamp-2 text-sm text-neutral-500">
                {p.product_description}
              </p>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
