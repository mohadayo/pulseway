import app from "./app";

const port = parseInt(process.env.DASHBOARD_PORT || "8002", 10);

app.listen(port, "0.0.0.0", () => {
  console.log(`${new Date().toISOString()} [INFO] dashboard: listening on port ${port}`);
});
