# Core flow: from need to match
1. HR adds a newcomer (Bogotá, arrives 15 May, furnished room) → `newcomers` row, stage `invited`, invite email sent.
2. Matching worker scores listings → inserts deduped `notifications` for agencies with listings in that city.
3. Agency feed shows "Newcomer from Company X · furnished · ≤ COP 2M · 92% match" with name/contact hidden.
4. No plan/credits → "Subscribe to unlock". With credit/plan → API returns contact; `contact_unlocks` logged; messaging enabled.
5. Newcomer/HR progress the stage to `contract_signed` → `moved_in`; review prompt sent.
