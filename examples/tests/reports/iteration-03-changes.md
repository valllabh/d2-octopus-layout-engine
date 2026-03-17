# Iteration 03: Changes Made

**Date**: 2026-03-17

## Changes

1. **Same column/row straight line enforcement**: When source and destination boxes are in the same grid column (centers within 5px X difference) and both anchors are vertical, the route is forced to a straight vertical line using the average X of the two anchor points. Same logic for same row horizontal edges. This prevents the anchor distribution from creating unnecessary bends on aligned edges.

2. **R5 rule clarification**: Added note to edge-routing-rules.md that D2 renders rounded corners at bend points which are cosmetic. Validators should count direction changes, not visual curves. A 3 point route has 1 bend regardless of D2's visual rendering.

3. **Validator guidance updated**: Clarified that for diagonal connections where obstacles exist between source and destination, 2 bends is the minimum possible and should be counted as PASS, not FAIL.

## Fixes Verified

- **test -> sec in CI/CD pipeline (28)**: Was 4 points / 2 bends because anchor distribution shifted X positions. Now forced to 3 points (straight vertical) since boxes are in the same column.

## Issues Confirmed as Correct (Not Bugs)

- **lint -> test in 28**: 2 bends because Security Scan at (2,3) blocks the direct L path from right of Lint to bottom of Test. 2 bends is the minimum given this obstacle. CORRECT behavior.
- **07 diamond pattern**: All 4 routes are mathematically L shapes (3 points, 1 bend). Debug logging confirmed. D2's rounded corner rendering makes them look like 2 bends visually. FALSE POSITIVE from validator.
- **rollback -> stage in 28**: Clean L shape (3 points, 1 bend). Goes left then up. CORRECT.
