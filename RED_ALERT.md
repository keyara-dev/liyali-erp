# RED_ALERT

This document tracks high-priority issues that block or weaken core platform behavior and need follow-up before they are treated as resolved.

## Current Item

- Organization suspension is not fully enforced yet. The admin API can mark an organization as suspended by setting `organizations.active = false`, but tenant access middleware does not explicitly block requests for suspended organizations everywhere. User suspension is enforced more directly through `users.active = false`.

- We need to come back to this and harden org-level suspension so it becomes a consistent access-control check across the platform.
