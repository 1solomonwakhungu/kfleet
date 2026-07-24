# Contra Live-Form Evidence Report

## Evidence boundary

- **Retrieved:** 2026-07-24 14:39-14:43 CDT (19:39-19:43 UTC).
- **Method:** Public Contra signup, first-party help pages, and Contra-hosted current signup assets. No account, emailed-code request, OAuth continuation, upload, identity/wallet flow, job application, contact, or purchase.
- **Rule:** Ordinary Contra onboarding is distinguished from the separate Contra Labs network.

## Source register

| ID | URL | Observable evidence |
|---|---|---|
| S1 | https://contra.com/sign-up | HTTP 200 public signup. |
| S2 | https://builds.contra.com/assets/assets/chunks/shared~entries/src/pages/log-in~entries/src/pages/sign-up-eHkcidon.js | Current signup implementation. |
| S3 | https://builds.contra.com/assets/assets/chunks/CodeInput-B3PAMfP6.js | Current six-character code control. |
| S4 | https://help.contra.com/en/articles/9322381-onboarding-and-completing-your-profile | Current independent onboarding/profile-completion guide. |
| S5 | https://help.contra.com/en/articles/9322393-how-to-build-a-case-study-from-scratch-on-contra | Current work-sample guide. |
| S6 | https://help.contra.com/en/articles/9322955-how-to-verify-your-identity-on-contra | Current Persona verification guide. |
| S7 | https://help.contra.com/en/articles/9322950-your-contra-wallet | Current wallet guide. |
| S8 | https://help.contra.com/en/articles/9322973-applying-to-jobs-on-contra | Current jobs guide. |
| S9 | https://help.contra.com/en/articles/11758871-how-to-update-your-availability-status-on-your-profile | Current availability guide. |
| S10 | https://help.contra.com/en/articles/14757116-applying-to-the-contra-labs-network | Separate Contra Labs application process. |
| S11 | https://contra.com/pricing | Current pricing page. |
| S12 | https://help.contra.com/en/articles/9322934-fees-overview | Current fees guide. |

## Directly observed signup

The live route offers **Continue with Google** or email signup with required **First name**, **Last name**, and **Email address**. Email uses HTML `type="email"` plus a dotted-domain validation pattern. Cloudflare Turnstile is configured. The next implementation state says **We emailed you a code** and accepts exactly six alphanumeric characters with a resend action. No password, GitHub signup, resume/CV, or upload field appears in the public signup. [S1][S2][S3]

## Current independent profile completion

After authentication, the current onboarding guide tells an Independent to select **Share work**, upload a profile photo, write a one-liner, choose Free or Pro, select topics, and complete the profile. A paid selection is offered but is not required. [S4]

Contra lists these requirements for profile completion and eligibility for client discovery/matched-job applications: [S4]

- Cover image or video.
- Four pieces of work.
- Hourly rate.
- At least one social or portfolio link; LinkedIn is recommended, not mandatory.
- Identity verification and wallet setup.

Availability is a separate profile setting with Available/Unavailable states, not listed in that completion checklist. [S9]

For each work sample, Contra documents image/video, work description, title, preview description, cover, tools, skills, client industry, optional collaborators/client/project duration/completed-work link, and publish/draft controls. Recommended dimensions are 1600 x 1200; the stated image/video maximum is 200 MB. The guide does not identify every metadata item as mandatory. [S5]

## Verification, jobs, and inaccessible facts

- Email signup requires the six-character emailed code. [S2][S3]
- Wallet setup requires identity verification. Contra redirects the user to Persona after the user selects the country issuing the ID; Persona collects personal details, ID information, and payment details. Exact accepted ID types, selfie/liveness requirements, country-specific payout fields, review times, rejection rules, and tax forms are inaccessible without starting that user-controlled flow. [S6][S7]
- No first-party source reviewed documents a general portfolio interview, test, or admission queue for an ordinary Independent account. This does not rule out internal moderation.
- `/jobs` and `/independent/discovery` redirected unauthenticated requests to login during retrieval. Exact job questions, attachments, proposal fields, rates, submission constraints, screening, and status tracking are auth-only. S8 says authenticated users can filter and choose Apply, Dismiss, or optionally Refer & Earn.
- Contra Labs separately uses one application, portfolio review, recorded video interview, and possible skills assessment. Those are not ordinary Contra requirements. [S10]

## User-only actions and exact gaps

Solomon must personally:

- Choose Google or email and complete authentication/code verification.
- Upload a truthful photo and cover; provide a one-liner, topics, and hourly rate.
- Produce and publish **four** genuine work samples with media. Existing copy contains only three text case-study outlines and no verified media assets.
- Provide a public social/portfolio URL.
- Select the ID-issuing country and complete Persona identity verification.
- Enter payout/payment details and complete wallet setup.
- Inspect each authenticated job's actual questions and decide whether to apply.

Every client name, project claim, metric, image, rate, and availability statement must be validated before publication.

## Pricing and safety boundary

The pricing page describes limited Free job access and displays $199/year or $29/month for Pro. A first-party help article elsewhere implies approximately $204/year, so paid pricing is internally inconsistent. Commission-free does not mean fee-free; S12 documents platform, processing, payout, and currency-conversion fees. No purchase is needed to prepare this packet, and this report makes no purchase recommendation. [S11][S12]

## Reconciliation of prior assumptions

| Prior claim | Current determination |
|---|---|
| Email, Google, or GitHub signup | **Partly incorrect.** Current public signup exposes email or Google, not GitHub. |
| Anyone can immediately start applying; no barriers | **Overbroad.** Jobs are auth-gated and matched-job eligibility requires profile completion. |
| No matching service | **Incorrect.** Contra documents matched jobs. [S4] |
| Profile requires bio, skills, experience, photo, portfolio, rate, availability | **Mixed.** Current completion list is cover, four works, rate, social link, identity, wallet; photo and one-liner appear in onboarding; availability is separate. |
| Free plan is fully usable | **Stale.** Current pricing says Free job access is limited and transaction-related fees exist. [S11][S12] |
| Pro costs exactly $199/year and should be considered | **Do not rely on or recommend.** First-party pricing conflicts and purchase is outside scope. |
| First-client timing and $100-$150/hour are known | **Unsupported by current primary application evidence.** |

## Current readiness

**Observed status: Not submitted.** A fourth work sample, media/cover assets, social or portfolio URL, and all user-controlled verification/wallet actions remain gaps.
