# Matching v1 (rule-based)
Score 0–100 per (newcomer, property). Implemented in `apps/api/src/matching.ts` and unit-tested.

| Factor | Points | Rule |
|---|---|---|
| Budget | 35 | ≤ max budget = 35; ≤ +10 % = 15; ≤ +25 % = 5; **> +25 % = reject (0)**. No budget set = 25 |
| Availability | 20 | available ≤ 14 days after arrival = 20; ≤ 30 days = 8; unknown = 20 |
| Bedrooms | 15 | ≥ wanted = 15; one fewer = 5 |
| Furnished | 15 | equal = 15; wants unfurnished but furnished = 10; no preference = 15 |
| District | 15 | in preferred list = 15; no preference = 8 |

City must match (case/accent-insensitive) or the score is 0. Agencies are notified when their **best** listing scores ≥ 50. Notifications are de-duplicated per (agency, newcomer) and withdrawn when the newcomer reaches `contract_signed`, `moved_in` or `cancelled`.
Next: learn weights from conversion data (see issue "AI matching v2").
