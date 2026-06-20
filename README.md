# 🚀 PROJECT BLUEPRINT: Chronos (Software Archaeology Simulator)

Chronos is a developer tool that reverse-engineers believable, auditable Git histories for completed software projects. Unlike simple "commit spammers," Chronos uses a non-destructive "Parallel Revamp" architecture. It mathematically deconstructs a final codebase and rebuilds it as a methodical, months-long enterprise development cycle, outputting a completely new, verified repository alongside the original.

## 1. THE "PARALLEL REVAMP" WORKFLOW (Crucial Concept)
Chronos must **never** alter, corrupt, or delete the user's original project files. The workflow is strictly non-destructive.

1.  **Source Selection:** The user selects their completed project directory.
2.  **Deep Scan:** The engine scans the Source Directory for folder size, file count, and Git data.
3.  **Configuration:** The user configures the desired timeline, AI settings, and realism injectors.
4.  **Output Selection:** The user selects an Output Directory (where the revamped project will be created).
5.  **Execution:**
    *   Creates the Output Directory and initializes a "Shadow Repo" inside.
    *   Generates dummy "skeleton" files and commits them over the simulated date range.
    *   **The Overlay Step:** Copies the user's actual, final project files from the Source Directory into the Output Directory.
    *   **Verification:** Runs `git status` in the Output Directory.
6.  **Result:** Original folder remains untouched, new folder contains the project with a pristine Git history.

## 2. TECHNICAL ARCHITECTURE (Zero JavaScript Allowed)
*   **Frontend (UI):** Flutter (Dart) Desktop Application.
*   **Backend (Engine):** Go (Golang).
*   **Communication (IPC):** JSON over STDIN/STDOUT (NDJSON logs).

## 3. ENGINE CONFIGURATION (The "Brain")
*   **Deterministic (Zero AI):** Pure Go logic and predefined templates.
*   **API-Driven (Unified Cloud/Local AI):** Standard HTTP REST API calls to an OpenAI-compatible endpoint.

## 4. REALISM INJECTORS (The Illusion of Fallibility)
*   AST Code Phasing
*   Dependency Timeline Alignment
*   Human Error Injection
*   Branch & PR Simulation
*   Chronobiometrics
*   Focus Directives

## 5. THE UI/UX: SINGLE-PAGE WORKSPACE
A clean, dark-mode, developer-centric layout.
- **Left Pane:** Configuration (Source/Output selection, Engine/Timeline config, Realism injectors).
- **Right Pane:** Live State & Logs (State summary, Live terminal logs, Verification status).

## 6. IPC PROTOCOL CONTRACT (Flutter <-> Go)
NDJSON format:
```json
{"type":"log","level":"info","message":"Scanning source directory..."}
{"type":"estimate","payload":{"commits":45,"branches":6,"runtime":"3m","size":"+2.4MB"}}
{"type":"state","payload":{"status":"completed","before":{"commits":2},"after":{"commits":45},"verified":true,"output_path":"/path/to/output_dir","report_path":"/path/to/output_dir/project_summary.md"}}
```

---

## 🛠 Build Instructions

### Prerequisites
- **Go** (1.24.3 or later)
- **Flutter** (3.41.2 or later)
- **Make**

### Building
```bash
make all
```
Ensure the `chronos-engine` binary is available for the UI.
