# Gun.io Live-Form Evidence Report

## Evidence boundary

- **Retrieved:** 2026-07-24 14:19-14:29 CDT (19:19-19:29 UTC).
- **Method:** Public Gun.io pages, unauthenticated application routes, and current Gun.io-served client assets/API. No account was created, email sent, form submitted, verification bypassed, purchase made, or client contacted.
- **Rule:** A field is called required only when the live UI/client validation or a current first-party page says so. Authenticated profile details are labeled inaccessible.

## Source register

| ID | URL | Observed result |
|---|---|---|
| S1 | https://app.gun.io/sign-up/ | HTTP 200, public interactive signup. |
| S2 | https://app.gun.io/site_media/static/assets/Signup-9fa8188f.js | Current Gun.io signup client asset. |
| S3 | https://app.gun.io/complete-sign-up/ | HTTP 200; without a valid emailed code, renders an invalid-code state and no usable form. |
| S4 | https://app.gun.io/site_media/static/assets/CompleteSignup-a349c4be.js | Current token-gated completion-form implementation. |
| S5 | https://app.gun.io/api/v2/marketing-attributions/ | Public current attribution choices. |
| S6 | https://gun.io/find-work/ | Public current talent-process page. |
| S7 | https://gun.io/faq/ | Public current screening, approval, rate, and hiring guidance. |
| S8 | https://gun.io/jobs/ | Public job summaries; applying routes to signup. |
| S9 | https://gun.io/candidate-interview-guide/ | Public candidate-interview guidance. |

## Directly observed entry flow

1. S1 first asks **What is your goal with Gun.io?** The talent choice is **To Find Work**.
2. The next screen asks the user to confirm **Join our talent network** and **Create talent profile**.
3. The public initial form exposes one required field: **Email**, using HTML `type="email"`. The client trims and lowercases it. [S1][S2]
4. Submission represents agreement to Gun.io's Terms of Use and Privacy Policy. A successful request displays **Check your email for a link to complete your signup!** [S1][S2]
5. No resume, GitHub, portfolio, rate, availability, work-history, or upload control appears before that email action. [S1][S2]

## Email-gated completion fields

The live completion route cannot be traversed without a valid emailed code. Its current Gun.io-served implementation defines the following fields, but they remain token-gated rather than directly usable in this review. [S3][S4][S5]

| Field | Observable rule |
|---|---|
| Set password | Required; minimum 12 characters. |
| Confirm password | Required; must match. |
| First name | Required; title-case validation. |
| Last name | Required; title-case validation. |
| Phone number | Required; exact accepted format/countries require a valid code. |
| How did you hear about us? | Required select. |
| Attribution detail | Conditionally required for sources requesting detail. |
| Terms acknowledgment | Required. |

The browser timezone is recorded automatically, not entered by the user. Current top-level attribution values include ad, event/conference, another site, referral, search engine, Gun.io tools/content, LinkedIn, X, ChatGPT, Claude, Gemini, Deel Talent, and Other. [S4][S5]

## Current published screening path

Gun.io currently describes this sequence: complete a talent profile with work history, preferred languages, and other information; reach **100% profile completeness**; pass initial/algorithmic screening; receive staff profile rating/review and work-history/background review; complete a technical approval call/live interview with a senior engineer or senior staff developer; then become eligible for shortlisting, Gun.io discussion, and client interview. Contracts and payment setup occur after client selection. [S6][S7][S9]

The newer jobs page summarizes onboarding as a quick screening, basic profile, and team intro. It does not prove the FAQ's more detailed stages have been removed. [S7][S8]

## Documents, gaps, and user-only actions

**No current public source establishes a mandatory resume/CV, GitHub, portfolio, code sample, certification, ID document, or reference upload.** Resume attention appears in a developer testimonial, not as a current required field. Exact profile fields, uploads and limits, rate format, availability controls, completeness formula, screening questions, and background-check consent UI are authentication-gated. [S1][S3][S6][S7]

Solomon must personally:

- Choose **To Find Work**, provide email, and request the confirmation link.
- Open the emailed link, create credentials, provide name/phone/attribution, and accept terms.
- Inspect and complete the authenticated profile to 100%, including only fields actually shown.
- Set and validate any requested rate and availability.
- Complete screening, approval call, shortlisting calls, and client interviews.
- Review and execute contracts if selected. Do not contact clients outside Gun.io before contract. [S7][S9]

## Reconciliation of prior assumptions

| Prior claim | Current determination |
|---|---|
| Public signup asks for work history, languages, resume, GitHub, rate, and availability | **Incorrect.** Public signup asks only for email; substantive profile fields are gated. |
| Profile must reach 100% | **Verified.** [S6] |
| Algorithmic screening, work-history/background review, and senior technical interview | **Verified at stage level.** Exact questions/checks remain inaccessible. [S7] |
| Resume and GitHub are mandatory | **Not established.** |
| Recommended rate is $100-$150/hour | **Unsupported by current first-party requirements.** Gun.io says developers receive their requested rate but publishes no standard rate. [S7] |
| Vetting takes 1-2 weeks | **Unsupported.** No current first-party duration was found. |
| First introduction occurs 13 days after approval | **Incorrect.** Gun.io says clients hire in an average 13 days after requesting a candidate, a different denominator. [S7] |
| About 100 developers are placed monthly | **Incorrect wording.** S7 says about 1,000 join monthly and about 100 are approved to work with clients. |
| No public job board | **Stale.** Public summaries exist at S8; application mechanics remain gated. |

## Current readiness

**Observed status: Not submitted.** Public requirements are reconciled, but account/email verification and all gated profile requirements remain user-only. Existing profile copy can seed the gated form, but every employer name, metric, rate, availability statement, and project claim must be verified by Solomon before use.
