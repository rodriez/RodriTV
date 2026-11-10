# Challenging Questions Reference

Question bank and techniques for the Product Owner interview during Discovery. Pick what fits; don't fire the whole list. Each challenge should come with a proposal or hypothesis (balanced-challenging stance).

## Table of contents

- Uncovering the real problem
- Challenging the requested solution
- Users and stakeholders
- Functionality and scope (v1 slicing)
- Testability / acceptance
- Success metrics
- Closing the interview

## Uncovering the real problem

- "What happens today when someone faces this problem? Walk me through the current workaround."
- "Who complains about this the most, and what do they say verbatim?"
- "If we did nothing for 6 months, what would break? What's the cost of the status quo?"
- "Is this the problem, or a symptom of a bigger one?" (apply 5 Whys when answers feel shallow)
- "How often does this happen? Daily pain or once-a-quarter annoyance?"

Technique: when the user opens with a feature request ("we need a dashboard"), respond with: "Before the dashboard — what decision can't you make today because you lack it?" Then propose: "My hypothesis is the real need is X; the dashboard is one way to serve it. Am I close?"

## Challenging the requested solution

- "You described a solution. What problem does it solve, in one sentence?"
- "What alternatives did you consider and discard? Why did they lose?"
- "If we could only ship half of this, which half?"
- "What would make this feature a failure even if we build it perfectly?"
- Counter-propose: offer a simpler or different approach and ask what it misses.

## Users and stakeholders

- "Who exactly is the first user? Name the role, not 'users' in general."
- "Who benefits but never touches the product? What do they expect from it?"
- "Who might resist this change internally?"

## Functionality and scope (v1 slicing)

- "What is the single action a user must be able to complete for this to be worth anything?"
- "Which of these requirements validates the core value? Which are comfort?"
- "If we ship v1 with only [smallest slice], would you approve it? If not, what's genuinely missing?"
- "What are we explicitly NOT doing in this version?" (capture for the out-of-scope section)

## Testability / acceptance

- "How will you personally verify this works, sitting in front of the product?"
- "Describe one concrete scenario: user does X, and you expect Y to happen."
- "What would make you reject the delivery on sight?"
- Keep it product-level: observable behavior and outcomes, never technical test types.

## Success metrics

- "What number moves if this succeeds? What's it at today, and what's the target?"
- "When would we check that metric — 30 days after release? A quarter?"
- "If the metric doesn't move, was the problem wrong or the solution?"
- Reject unmeasurable goals ("improve experience") — ask: "How would we notice the improvement without talking to anyone?"

## Closing the interview

- "What's the riskiest assumption in everything we just said?"
- "What question should I have asked and didn't?"
- Confirm the v1 slice, the metrics, and the out-of-scope list before drafting the PRD.
