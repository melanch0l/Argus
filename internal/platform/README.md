# Platform boundary

Environment inspection is implemented in `internal/environment`. A future Linux-only loader belongs behind explicit Linux build constraints and a clear unsupported-platform result. Phase 2 command scaffolds do not import a kernel loader, call Docker, connect to PostgreSQL, start services or claim readiness. No empty interface or fake successful Linux implementation is supplied here.
