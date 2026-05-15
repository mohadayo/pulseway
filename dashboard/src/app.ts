import express, { Request, Response } from "express";

const app = express();
app.use(express.json());

const startTime = Date.now();

const LOG_LEVEL = process.env.DASHBOARD_LOG_LEVEL || "info";

function log(level: string, message: string): void {
  if (level === "debug" && LOG_LEVEL !== "debug") return;
  const ts = new Date().toISOString();
  console.log(`${ts} [${level.toUpperCase()}] dashboard: ${message}`);
}

interface DashboardConfig {
  name: string;
  widgets: string[];
  createdAt: string;
}

const dashboards: Map<string, DashboardConfig> = new Map();

app.get("/health", (_req: Request, res: Response) => {
  res.json({
    status: "healthy",
    service: "dashboard",
    uptime_seconds: Math.round((Date.now() - startTime) / 1000),
    dashboard_count: dashboards.size,
  });
});

app.get("/dashboards", (_req: Request, res: Response) => {
  log("info", `Listing ${dashboards.size} dashboards`);
  res.json({ dashboards: Array.from(dashboards.values()) });
});

app.post("/dashboards", (req: Request, res: Response) => {
  const { name, widgets } = req.body;
  if (!name || typeof name !== "string") {
    log("error", "Dashboard creation failed: missing name");
    res.status(400).json({ error: "name is required and must be a string" });
    return;
  }
  if (dashboards.has(name)) {
    log("error", `Dashboard creation failed: '${name}' already exists`);
    res.status(409).json({ error: `dashboard '${name}' already exists` });
    return;
  }
  const config: DashboardConfig = {
    name,
    widgets: Array.isArray(widgets) ? widgets : [],
    createdAt: new Date().toISOString(),
  };
  dashboards.set(name, config);
  log("info", `Created dashboard '${name}' with ${config.widgets.length} widgets`);
  res.status(201).json(config);
});

app.get("/dashboards/:name", (req: Request, res: Response) => {
  const config = dashboards.get(req.params.name);
  if (!config) {
    log("debug", `Dashboard '${req.params.name}' not found`);
    res.status(404).json({ error: "dashboard not found" });
    return;
  }
  log("info", `Fetched dashboard '${req.params.name}'`);
  res.json(config);
});

app.delete("/dashboards/:name", (req: Request, res: Response) => {
  if (!dashboards.has(req.params.name)) {
    res.status(404).json({ error: "dashboard not found" });
    return;
  }
  dashboards.delete(req.params.name);
  log("info", `Deleted dashboard '${req.params.name}'`);
  res.json({ deleted: true });
});

export default app;
