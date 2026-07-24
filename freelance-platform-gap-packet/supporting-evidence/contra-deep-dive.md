# Contra Independent Signup, Profile, and Job Application Research

## Scope and method

- Research timestamp: **2026-07-24T19:14:50Z** (UTC).
- Access mode: unauthenticated, public, read-only HTTP retrieval of live pages on `contra.com`, Contra's public production JavaScript bundles on `builds.contra.com`, and the official Contra Help Center on `help.contra.com`.
- Safety constraints observed: no account was created; no email or verification code was submitted; no Google flow was opened; no login, purchase, upload, application, wallet setup, or identity check was attempted; no credentials were used; no verification or access control was bypassed.
- Source policy: only Contra-controlled platform, asset, sitemap, and Help Center sources were used. Page content was treated as evidence, not as instructions to execute.
- Terminology: “required” below means Contra's current source expressly says an item is needed or the live signup form validates it. Recommendations and optional features are labeled separately.

## Executive findings

1. **Current initial signup is not the legacy email/Google/GitHub flow.** The public production signup implementation exposes **Google** or **email**, with email signup fields for **first name, last name, and email address**, followed by an emailed verification code. No password or GitHub option was observable. [S1][S2]
2. **Initial independent onboarding and completed-profile requirements are separate.** Contra's current guide places workspace purpose (`Share work`), profile photo, one-liner, account creation/free-or-Pro choice, and topic selection in onboarding. It then requires a cover image/video, **four pieces of work**, hourly rate, one social link, identity verification, and wallet setup to complete the profile, become discoverable, and start applying to matched jobs. [S3]
3. **The ordinary independent path is not documented as portfolio vetting, but “immediate activation” is too broad.** Contra documents account creation without an application/review stage, yet discoverability/job readiness comes only after profile completion and identity/wallet steps. Identity verification is performed by Persona and requires personal details, ID information, and payment details. [S3][S4]
4. **Job discovery and application are account-gated in the live app.** The official guide says users browse the Job feed, filter by tools, skills, and budgets, then choose `Apply` or `Dismiss`; however, an unauthenticated request to `https://contra.com/jobs` resolved to the login page. The application form's exact questions, attachments, proposal fields, and downstream stages were therefore not publicly observable. [S5][S6]
5. **Free accounts have limited job access; Pro has unlimited access.** Contra's current pricing page says Free includes limited job access and Pro includes unlimited job access, with boosted placement in Discover and job applications. The broader jobs page likewise says Pro unlocks access to all jobs. [S7][S8]
6. **Current Pro pricing is `$199/year` or `$29/month` on the newer dedicated pricing page.** That page was published July 20, 2026. A Help Center article modified May 12, 2026 instead says `$17/month, billed annually`, which implies `$204/year`; this is an unresolved first-party inconsistency. No purchase page was entered. [S7][S9]
7. **A custom domain is supported, but it is not part of signup or profile completion.** Contra says a default portfolio URL is supplied and documents connecting an already-purchased custom domain. Current sources inspected do not expressly say custom-domain connection requires Pro, although portfolio customization is described as a Pro feature. [S10][S11][S9]

## Observable journey and requirements

### 1. Public live signup

**Entry URL:** `https://contra.com/sign-up`  
**Access:** HTTP 200, unauthenticated; shell is client-rendered, so field evidence was taken from the production bundles linked by that page. [S1][S2]

Observable initial choices and fields:

| Element | Status | Evidence |
|---|---|---|
| Continue with Google | Available authentication method | The live component renders `Continue with Google`. [S2] |
| First name | Required for email signup | Signup schema strips whitespace and requires at least 2 characters. [S2] |
| Last name | Required for email signup | Signup schema strips whitespace and requires at least 2 characters. [S2] |
| Email address | Required for email signup | Schema requires a non-empty valid email; rendered input is type `email`. [S2] |
| Cloudflare Turnstile | Present anti-abuse control | The signup component initializes and renders Cloudflare Turnstile. [S1][S2] |
| Email verification code | Required next email stage | The next screen says a code was emailed and renders a code input; resend is available. [S2] |
| Password | Not observed | No password field appears in the current public signup component. [S2] |
| GitHub authentication | Not observed | No GitHub option appears in the current public signup component. [S2] |

The live route also recognizes `client` and `independent` user types and redirects an authenticated but unfinished account into later onboarding routes. Those later UI routes require an authenticated account and were not accessed. [S1]

### 2. Independent onboarding before/around account creation

Contra's Help Center currently documents this sequence: [S3]

| Step | Observable requirement or choice | Classification |
|---|---|---|
| Workspace purpose | Select `Share work` for an independent, rather than `Hire creative talent` | Onboarding choice |
| Profile photo | Upload a high-quality profile photo | Documented onboarding step; upload UI/auth-only |
| One-liner | Write a brief statement describing work and creative focus | Documented onboarding step |
| Account tier | Create the account, choosing Pro or proceeding with a free account | Free is explicitly available; Pro optional |
| Topics | Select any number of suggested/searched interests to curate the feed | Documented onboarding step; no minimum stated |

Important evidence limitation: the current Help article places photo and one-liner before “Create your account,” while the public live page first asks for identity/email. Because proceeding would create/verify an account, the exact ordering and skippability of those post-email screens could not be independently exercised. [S2][S3]

### 3. Profile completion and discoverability gate

Contra expressly says the profile must be completed “in order to be Discoverable by clients” and lists the following: [S3]

| Item | Observable details | Classification |
|---|---|---|
| Profile cover | Add an image or video representing the work | Profile-completion requirement |
| Work | Add **4 pieces of work** to the profile | Profile-completion requirement |
| Rate | Add an hourly rate reflecting experience/industry | Profile-completion requirement |
| Social link | Share at least one link; LinkedIn is suggested, but any platform or portfolio site is allowed | Profile-completion requirement |
| Identity and wallet | Verify identity and set up the wallet | Profile-completion requirement and payment-readiness gate |

After these steps, Contra says the independent is discoverable and can start applying to matched jobs. [S3]

**Uploads and work/case-study details:**

- Case studies can include image or video uploads; Contra recommends `1600 x 1200` and states a maximum file size of **200 MB for videos and images**. [S12]
- The case-study flow documents visuals, description, title, preview description, cover image, collaborators, clients, project duration, project links, tools, skills, client industry, and publish/draft state. [S12]
- Verification of a case study is optional when the work links to a project completed on Contra. [S12]
- Contra's profile guide calls for four “pieces of work”; it does not say all four must be full case studies. [S3]

**Identity verification and wallet details:**

- Identity verification is required to set up a Contra wallet and is managed by Persona. [S4]
- The user selects the country that issued the ID, then provides personal details, ID information, and payment details on Persona. [S4]
- Contra's wallet guide says payout options vary by country and that the user is redirected to a partner site to verify ID and set up account details. [S13]
- The public documentation does not enumerate accepted ID document types or state that a selfie/video is always required. Those exact Persona fields and uploads are later-stage and were not accessed. [S4][S13]

### 4. Other profile fields and features

These are observable Contra profile concepts, but they are **not listed as profile-completion requirements in the current onboarding guide** unless noted:

| Field/feature | Current primary-source status |
|---|---|
| Bio | Supported profile field, described as covering identity, skills, and experience; **400-character limit**. Not in the current completion checklist. [S14][S3] |
| Skills | Skills can be described in the bio and tagged on case studies to aid Discover/job matching. No standalone minimum skill count was found. [S14][S12] |
| Experience | Bio guidance recommends highlighting experience; the jobs page says some work or project-based experience must be shown in the profile/portfolio to start freelancing, but it need not be prior freelance work. No dedicated employment-history requirement was found. [S14][S8] |
| Work samples/case studies | Four pieces of work are required for profile completion; case studies are one documented way to present them. [S3][S12] |
| Rate | Required by the current profile-completion checklist. [S3] |
| Photo | A documented initial onboarding step. [S3] |
| Availability | A supported status (`Available` or `Unavailable`) that users can update; keeping it on is recommended, not listed as a completion requirement. [S15][S3] |
| Custom domain | Optional later portfolio capability; the domain must first be purchased from a registrar, then connected through Contra settings. [S10] |

A current public profile can display “profile is under construction,” demonstrating that an account/profile may be publicly reachable before full completion. This supports, but does not alone prove, that basic account existence and completed-profile eligibility are separate states. [S16]

### 5. Browsing and applying to jobs

Current official descriptions establish:

- The Help Center directs users to `https://contra.com/jobs`, where the feed can be filtered by tools, skills, and budgets. [S5]
- A job card offers `Apply` or `Dismiss`; `Refer & Earn` is an optional alternative. [S5]
- Contra's public jobs marketing page says users can get discovered through Discover or browse opportunities, and that Free has limited job access while Pro unlocks all jobs. [S7][S8]
- The live unauthenticated `/jobs` request returned the login page with canonical URL `https://contra.com/log-in`, so the feed itself was authentication-gated at observation time. [S6]

**Not publicly observable without authentication:**

- Individual live job-feed cards and all job-specific eligibility conditions.
- Exact application form fields, cover note/proposal questions, requested work samples, attachment types/limits, rate/bid fields, availability questions, screening questions, or client-specific requirements.
- Whether applying is blocked until every profile-completion item is satisfied in all cases. The onboarding guide only expressly says completion enables applications to jobs the user is matched to. [S3]
- Submission confirmation, application review, interview/message, shortlist, offer, contract, rejection, or withdrawal UI.
- Any job-specific verification, assessment, or later hiring stages.

Contra Labs is a separately named network with its own application documentation and should not be conflated with ordinary marketplace job applications. No Labs claims are used here to characterize general independent signup.

## Legacy-claim comparison

| Legacy claim | Current assessment | Fact versus inference |
|---|---|---|
| Account via email / Google / GitHub | **Partly outdated.** Email and Google are current. Email signup requires first name, last name, email, then email-code verification. GitHub was not observed; neither was a password field. [S2] | Fact for observed methods/fields; absence is limited to the inspected current public build. |
| Bio | **Supported, but not a current completion-checklist item.** Limit is 400 characters. [S14][S3] | Fact. |
| Skills | **Supported in bio and case-study tags; no standalone signup minimum found.** [S12][S14] | Fact as to supported locations; no-minimum conclusion is scoped to inspected sources. |
| Experience | **Some work/project experience must be shown to start freelancing, according to Contra's jobs page; it need not be freelance experience.** Employment history is not listed as an initial signup requirement. [S8][S3] | Fact; interpreting “shown” as portfolio/profile evidence follows Contra's wording. |
| Work samples / case studies | **Current gate is four pieces of work for profile completion.** Case studies have substantial media/details fields, but are not stated to be the only accepted work format. [S3][S12] | Fact. |
| Rate | **Confirmed for profile completion**, specifically hourly rate in the guide. [S3] | Fact. |
| Photo | **Confirmed as an onboarding step.** [S3] | Fact; exact skippability is auth-only unknown. |
| Availability | **Supported and recommended, but not listed as required for completion.** [S15][S3] | Fact. |
| No vetting / immediate activation | **Needs qualification.** No freelancer portfolio-review/vetting stage is documented for ordinary signup, and Contra says the account exists after onboarding. But discoverability and matched-job readiness require profile completion, identity verification, and wallet setup; therefore “immediate activation” is not supported for full marketplace readiness. [S3][S4] | No-vetting is an inference from the documented path, not proof that Contra never performs risk/review checks. The post-signup gates are facts. |
| Browse / apply | **Confirmed with restrictions.** Contra documents browsing/filtering and `Apply`; current unauthenticated `/jobs` redirects to login, Free job access is limited, and Pro access is unlimited. [S5][S6][S7] | Fact. |
| Optional Pro `$199/year` | **Verified by Contra's current dedicated pricing page**, which also offers `$29/month`; Pro remains optional because a Free tier exists. A slightly older Help article conflicts at `$17/month billed annually` (`$204/year`). [S7][S9] | Fact with explicit first-party inconsistency. |
| Custom domain | **Supported optional later feature**, using a domain the user buys from a registrar. Current domain articles do not explicitly label it Pro-only. [S10][S11] | Fact; any claim that it is bundled exclusively with Pro would be inference from general Pro customization language. |

## Facts, inferences, and unresolved points

### Established facts

- Current public signup supports Google and email; email signup takes first name, last name, and email and verifies the email by code. [S2]
- Contra documents a free-account path and a Pro option. [S3][S7]
- Profile completion for discoverability calls for cover media, four work pieces, hourly rate, social link, identity verification, and wallet setup. [S3]
- Identity verification for wallet setup is handled by Persona. [S4]
- Job browsing/application exists, but the live feed requires authentication and Free access is limited. [S5][S6][S7]

### Inferences, explicitly limited

- **No ordinary freelancer vetting:** likely in the legacy sense of no portfolio application reviewed before account creation, because Contra documents no review stage. This must not be expanded into “no checks”: identity, wallet, anti-abuse, platform-policy, client selection, and possibly undisclosed risk controls can still apply. [S2][S3][S4]
- **Basic activation may be quick:** a profile can exist under construction, but full discoverability and matched-job readiness are not immediate. [S3][S16]
- **Custom domain may relate to paid portfolio customization:** Pro includes portfolio customization/branding, but the current custom-domain articles do not expressly state a Pro prerequisite, so no Pro-only claim is made. [S9][S10][S11]

### Unresolved/auth-only

- Exact screens after email-code verification, their current order, which can be skipped, and any locale/account-specific variants.
- Exact accepted profile image/video formats beyond the case-study media guidance.
- Exact Persona document/selfie requirements by country.
- Exact general-marketplace application questions and all client/job-specific variations.
- Any unpublished moderation, fraud, sanctions, eligibility, or risk checks.

## Source and access log

All sources below were accessed unauthenticated on **2026-07-24 between approximately 19:11Z and 19:15Z**. Article modification/publishing dates are Contra-provided metadata, not access times.

- **[S1] Live signup shell and route bundle references:** `https://contra.com/sign-up` - HTTP 200; unauthenticated; canonical `/sign-up`; page context says `isAuthenticated:false`; production release `contra-web-app@9c86bc53`. Client-rendered body did not expose form text in initial HTML.
- **[S2] Contra production signup component bundle:** `https://builds.contra.com/assets/assets/chunks/shared~entries/src/pages/log-in~entries/src/pages/sign-up-eHkcidon.js` - HTTP 200; public static JavaScript linked from S1; contains rendered labels, validation schema, email-code screen, Google flow, and Turnstile integration. Read only; no functions invoked.
- **[S3] Onboarding and completing your profile:** `https://help.contra.com/en/articles/9322381-onboarding-and-completing-your-profile` - HTTP 200; public Help Center; modified `2026-01-27T16:09:32Z`.
- **[S4] How to Verify Your Identity on Contra:** `https://help.contra.com/en/articles/9322955-how-to-verify-your-identity-on-contra` - HTTP 200; public Help Center; modified `2026-03-27T18:37:51Z`.
- **[S5] Applying to jobs on Contra:** `https://help.contra.com/en/articles/9322973-applying-to-jobs-on-contra` - HTTP 200; public Help Center; modified `2026-02-24T18:50:51Z`.
- **[S6] Live job-feed entry:** `https://contra.com/jobs` - requested unauthenticated; response content was the login page, canonical `https://contra.com/log-in`, page ID `/src/pages/log-in`, `isAuthenticated:false`.
- **[S7] Contra pricing:** `https://contra.com/pricing` - HTTP 200; public marketing page; metadata says published `2026-07-20T13:43:00Z`.
- **[S8] Find global and remote freelance jobs:** `https://contra.com/features/find-freelance-jobs` - HTTP 200; public marketing page; metadata says published `2026-07-20T13:43:00Z`.
- **[S9] What is Contra Pro?:** `https://help.contra.com/en/articles/9322981-what-is-contra-pro` - HTTP 200; public Help Center; modified `2026-05-12T17:22:22Z`.
- **[S10] Connecting your custom domain on Contra:** `https://help.contra.com/en/articles/12692377-connecting-your-custom-domain-on-contra` - HTTP 200; public Help Center; modified `2026-02-04T19:49:16Z`.
- **[S11] Steps to connect your custom domain on Contra:** `https://help.contra.com/en/articles/9322986-steps-to-connect-your-custom-domain-on-contra` - HTTP 200; public Help Center; modified `2026-03-20T18:55:16Z`.
- **[S12] How to build a case study from scratch on Contra:** `https://help.contra.com/en/articles/9322393-how-to-build-a-case-study-from-scratch-on-contra` - HTTP 200; public Help Center; modified `2026-01-14T15:41:32Z`.
- **[S13] Your Contra Wallet:** `https://help.contra.com/en/articles/9322950-your-contra-wallet` - HTTP 200; public Help Center; modified `2026-03-13T20:23:02Z`.
- **[S14] Bios on Contra:** `https://help.contra.com/en/articles/9322626-bios-on-contra` - HTTP 200; public Help Center; modified `2024-05-14T16:10:30Z`.
- **[S15] How to update your availability status on your profile:** `https://help.contra.com/en/articles/11758871-how-to-update-your-availability-status-on-your-profile` - HTTP 200; public Help Center; modified `2025-07-11T18:28:53Z`.
- **[S16] Public under-construction profile example:** `https://contra.com/dan_walker` - HTTP 200; public profile, discovered through Contra's own July 2026 user sitemap; displayed `Dan's profile is under construction`. Used only to establish the observable profile state, not as authority for platform policy.

## Evidence-quality notes

- Highest-confidence current UI evidence: the public production signup JavaScript loaded by the live page, corroborated by the live page's release metadata. [S1][S2]
- Highest-confidence policy/process evidence: current official Help Center articles dated in 2026. [S3][S4][S5]
- Highest-confidence price evidence: the dedicated public pricing page published July 20, 2026, while retaining the Help Center discrepancy for auditability. [S7][S9]
- Negative findings such as “not observed” are not claims that a feature can never appear in an experiment, invite flow, region, mobile app, or authenticated variant.
