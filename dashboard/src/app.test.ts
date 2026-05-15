import request from "supertest";
import app from "./app";

describe("Dashboard API", () => {
  describe("GET /health", () => {
    it("returns healthy status", async () => {
      const res = await request(app).get("/health");
      expect(res.status).toBe(200);
      expect(res.body.status).toBe("healthy");
      expect(res.body.service).toBe("dashboard");
      expect(res.body).toHaveProperty("uptime_seconds");
      expect(res.body).toHaveProperty("dashboard_count");
    });
  });

  describe("POST /dashboards", () => {
    it("creates a dashboard", async () => {
      const res = await request(app)
        .post("/dashboards")
        .send({ name: "test-dash", widgets: ["cpu", "memory"] });
      expect(res.status).toBe(201);
      expect(res.body.name).toBe("test-dash");
      expect(res.body.widgets).toEqual(["cpu", "memory"]);
      expect(res.body).toHaveProperty("createdAt");
    });

    it("rejects duplicate name", async () => {
      await request(app).post("/dashboards").send({ name: "dup-dash" });
      const res = await request(app).post("/dashboards").send({ name: "dup-dash" });
      expect(res.status).toBe(409);
    });

    it("rejects missing name", async () => {
      const res = await request(app).post("/dashboards").send({});
      expect(res.status).toBe(400);
      expect(res.body.error).toContain("name is required");
    });
  });

  describe("GET /dashboards", () => {
    it("lists dashboards", async () => {
      const res = await request(app).get("/dashboards");
      expect(res.status).toBe(200);
      expect(Array.isArray(res.body.dashboards)).toBe(true);
    });
  });

  describe("GET /dashboards/:name", () => {
    it("returns 404 for unknown dashboard", async () => {
      const res = await request(app).get("/dashboards/nonexistent");
      expect(res.status).toBe(404);
    });

    it("returns existing dashboard", async () => {
      await request(app).post("/dashboards").send({ name: "fetch-me" });
      const res = await request(app).get("/dashboards/fetch-me");
      expect(res.status).toBe(200);
      expect(res.body.name).toBe("fetch-me");
    });
  });

  describe("DELETE /dashboards/:name", () => {
    it("deletes existing dashboard", async () => {
      await request(app).post("/dashboards").send({ name: "del-me" });
      const res = await request(app).delete("/dashboards/del-me");
      expect(res.status).toBe(200);
      expect(res.body.deleted).toBe(true);
    });

    it("returns 404 for unknown dashboard", async () => {
      const res = await request(app).delete("/dashboards/no-such");
      expect(res.status).toBe(404);
    });
  });
});
