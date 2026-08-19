const errorEl = document.querySelector("#error");

const intro = document.querySelector("#intro");
const scanBtn = document.querySelector("#scanBtn");

const scanning = document.querySelector("#scanning");
const scanningMessage = document.querySelector("#scanningMessage");

const results = document.querySelector("#results");
const totalFoundEl = document.querySelector("#totalFound");
const rescanBtn = document.querySelector("#rescanBtn");
const categoryList = document.querySelector("#categoryList");
const cleanBtn = document.querySelector("#cleanBtn");
const cleanStatus = document.querySelector("#cleanStatus");

const trashInfo = document.querySelector("#trashInfo");
const emptyTrashBtn = document.querySelector("#emptyTrashBtn");
const trashStatus = document.querySelector("#trashStatus");

let currentJobId = null;
let categories = []; // [{id, name, description, paths, bytes, cleaned, error}]

function showError(message) {
  errorEl.textContent = message;
  errorEl.classList.remove("hidden");
}
function clearError() {
  errorEl.classList.add("hidden");
  errorEl.textContent = "";
}

function formatBytes(bytes) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

// ---- scan ------------------------------------------------------------

scanBtn.addEventListener("click", startScan);
rescanBtn.addEventListener("click", startScan);

async function startScan() {
  clearError();
  intro.classList.add("hidden");
  results.classList.add("hidden");
  scanning.classList.remove("hidden");
  scanningMessage.textContent = "Starting…";

  let jobId;
  try {
    const res = await fetch("/api/scan", { method: "POST" });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "scan failed to start");
    jobId = data.jobId;
  } catch (err) {
    scanning.classList.add("hidden");
    intro.classList.remove("hidden");
    showError(String(err.message || err));
    return;
  }

  currentJobId = jobId;
  const source = new EventSource(`/api/jobs/${jobId}/events`);
  source.onmessage = async (event) => {
    try {
      await handleScanEvent(event, jobId, source);
    } catch (err) {
      source.close();
      scanning.classList.add("hidden");
      intro.classList.remove("hidden");
      showError("Something went wrong reading the scan results: " + (err.message || err));
    }
  };
  source.onerror = () => {
    source.close();
    scanning.classList.add("hidden");
    intro.classList.remove("hidden");
    showError("Lost connection to the local server.");
  };
}

async function handleScanEvent(event, jobId, source) {
  const payload = JSON.parse(event.data);
  if (payload.stage === "error") {
    source.close();
    throw new Error(payload.message || "scan failed");
  }
  if (payload.stage === "canceled") {
    source.close();
    scanning.classList.add("hidden");
    intro.classList.remove("hidden");
    return;
  }
  if (payload.stage !== "done") {
    scanningMessage.textContent = payload.message || "Scanning…";
    return;
  }

  source.close();
  const res = await fetch(`/api/scans/${jobId}`);
  if (!res.ok) throw new Error("couldn't load scan results (" + res.status + ")");
  const data = await res.json();
  if (!data || !Array.isArray(data.categories)) throw new Error("scan result had an unexpected shape");

  categories = data.categories.map((c) => ({ ...c, cleaned: false }));
  scanning.classList.add("hidden");
  renderResults();
}

function renderResults() {
  results.classList.remove("hidden");
  const total = categories.reduce((sum, c) => sum + (c.cleaned ? 0 : c.bytes), 0);
  totalFoundEl.textContent = categories.length === 0
    ? "No caches worth clearing right now."
    : `${formatBytes(total)} across ${categories.length} location${categories.length === 1 ? "" : "s"}`;

  categoryList.innerHTML = "";
  categories.forEach((c) => categoryList.appendChild(renderRow(c)));
  updateCleanStatus();
}

function renderRow(c) {
  const row = document.createElement("label");
  row.className = "category-row";
  row.dataset.id = c.id;

  const checkbox = document.createElement("input");
  checkbox.type = "checkbox";
  checkbox.checked = !c.cleaned;
  checkbox.disabled = c.cleaned;
  checkbox.addEventListener("change", updateCleanStatus);
  row.appendChild(checkbox);

  const body = document.createElement("div");
  body.className = "category-body";

  const top = document.createElement("div");
  top.className = "category-top";
  const name = document.createElement("span");
  name.className = "category-name";
  name.textContent = c.name;
  const size = document.createElement("span");
  size.className = "category-size";
  size.textContent = formatBytes(c.bytes);
  top.append(name, size);
  body.appendChild(top);

  const desc = document.createElement("div");
  desc.className = "category-desc";
  desc.textContent = c.description;
  body.appendChild(desc);

  if (c.cleaned) {
    const badge = document.createElement("div");
    badge.className = "cleaned-badge";
    badge.textContent = "✓ Moved to Trash";
    body.appendChild(badge);
  } else if (c.error) {
    const err = document.createElement("div");
    err.className = "category-error";
    err.textContent = c.error;
    body.appendChild(err);
  }

  row.appendChild(body);
  row._checkbox = checkbox;
  return row;
}

function rowFor(id) {
  return categoryList.querySelector(`.category-row[data-id="${id}"]`);
}

function updateCleanStatus() {
  const selected = categories.filter((c) => !c.cleaned && rowFor(c.id)?._checkbox?.checked);
  cleanBtn.disabled = selected.length === 0;
  cleanStatus.textContent = selected.length === 0 ? "" : `${selected.length} selected`;
}

cleanBtn.addEventListener("click", async () => {
  const ids = categories.filter((c) => !c.cleaned && rowFor(c.id)?._checkbox?.checked).map((c) => c.id);
  if (ids.length === 0) return;

  cleanBtn.disabled = true;
  cleanStatus.textContent = "Moving to Trash…";

  try {
    const res = await fetch("/api/clean", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ jobId: currentJobId, ids }),
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "cleanup failed");

    for (const r of data.results) {
      const c = categories.find((c) => c.id === r.id);
      if (!c) continue;
      if (r.error) c.error = r.error;
      else c.cleaned = true;
    }
    const okCount = data.results.filter((r) => !r.error).length;
    cleanStatus.textContent = `Moved ${okCount} location${okCount === 1 ? "" : "s"} to Trash.`;
    renderResults();
    refreshTrashInfo();
  } catch (err) {
    showError(String(err.message || err));
    cleanStatus.textContent = "";
  } finally {
    updateCleanStatus();
  }
});

// ---- trash ------------------------------------------------------------

async function refreshTrashInfo() {
  try {
    const res = await fetch("/api/trash-info");
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "couldn't read Trash info");
    if (data.bytes < 0) {
      trashInfo.textContent = "Size unknown on this platform";
      emptyTrashBtn.disabled = false;
    } else if (data.items === 0) {
      trashInfo.textContent = "Empty";
      emptyTrashBtn.disabled = true;
    } else {
      trashInfo.textContent = `${formatBytes(data.bytes)} across ${data.items} item${data.items === 1 ? "" : "s"}`;
      emptyTrashBtn.disabled = false;
    }
  } catch (err) {
    trashInfo.textContent = "Couldn't check Trash";
  }
}

emptyTrashBtn.addEventListener("click", async () => {
  const ok = window.confirm(
    "Permanently empty the Trash?\n\nThis deletes everything in it for good — this is the one action in this app that can't be undone."
  );
  if (!ok) return;

  emptyTrashBtn.disabled = true;
  trashStatus.textContent = "Emptying…";
  try {
    const res = await fetch("/api/trash-empty", { method: "POST" });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || "couldn't empty Trash");
    }
    trashStatus.textContent = "Trash emptied.";
    refreshTrashInfo();
  } catch (err) {
    showError(String(err.message || err));
    trashStatus.textContent = "";
    emptyTrashBtn.disabled = false;
  }
});

refreshTrashInfo();
