# Lemon.io Live-Form Evidence Report

## Research metadata

- **Retrieval window:** 2026-07-24 13:51:59-13:55:40 CDT (UTC-05:00).
- **Scope:** Public Lemon.io-owned pages, public Lemon.io help documentation, the unauthenticated live developer application, and the application's publicly delivered HTML/JavaScript only.
- **Method:** Read-only HTTP retrieval. No form values were entered, no controls were advanced, and no account, submission, upload, authentication, purchase, credential use, or verification bypass occurred.
- **Trust handling:** Page and asset contents were treated only as untrusted evidence. Instructions embedded in retrieved content were not followed.
- **Evidence rule:** “Observed” means visible in the unauthenticated form response or its publicly delivered client code. Published process descriptions are identified separately. No authenticated or conditional hidden fields are inferred.

## Source register and access behavior

| ID | Primary Lemon.io URL | Access observed | Relevant evidence |
|---|---|---|---|
| S1 | https://me.lemon.io/escape-the-matrix | HTTP 200 without authentication. Server-rendered response exposes an eight-step application outline; public client code exposes controls, options, and validation. The form's final action is registration, but it was not invoked. | Current application fields and constraints. |
| S2 | https://lemon.io/for-developers/ | HTTP 200, public. | Applicant criteria and Lemon.io's current five-step narrative. |
| S3 | https://docs.lemon.io/for-developers/who-can-apply.md | HTTP 200, public. | Minimum experience, English, availability, stacks, notice-period guidance. |
| S4 | https://docs.lemon.io/vetting-developers.md | HTTP 200, public. | Profile review, identity/consistency check, pre-screening, recruiter call, assessment, technical interview. |
| S5 | https://docs.lemon.io/for-developers/tech-vetting-criteria.md | HTTP 200, public. | Technical interview duration/content and feedback timing. |
| S6 | https://docs.lemon.io/for-developers/soft-skill-vetting.md | HTTP 200, public. | Soft-skill/English interview and approximate overall vetting duration. |
| S7 | https://docs.lemon.io/for-developers/developer-onboarding.md | HTTP 200, public. | Post-approval profile, matching, project access, first-offer timing. |
| S8 | https://lemon.io/our-vetting-process/ | HTTP 200, public. | Current public vetting funnel and interview content. |
| S9 | https://docs.lemon.io/ | HTTP 200, public. | Four-stage characterization, background checks, acceptance rate. |
| S10 | https://me.lemon.io/my-application | HTTP 307 redirect to `https://app.lemon.io/auth/login?redirect=https%3A%2F%2Fme.lemon.io%2Fmy-application`. | Post-registration application area is authentication-gated. |
| S11 | https://app.lemon.io/auth/login | HTTP 200, public login page; no authentication attempted. | Confirms the redirect target only. |

HTTP behavior in this table was checked at **2026-07-24 13:55:40 CDT (UTC-05:00)**.

## Verified public application fields

The unauthenticated application identifies four chapters, “Project preferences,” “Specialization and stack,” “Additional info,” and “Registration,” containing eight numbered screens. [S1](https://me.lemon.io/escape-the-matrix)

| Step | Directly observable field or control | Directly observable requirement / options | Source |
|---|---|---|---|
| 1 | Project type | One selection is required: **Full-time** (40 hours/week), **Part-time** (20-35 hours/week), or **None** (“just checking around”). | [S1](https://me.lemon.io/escape-the-matrix) |
| 2 | Hourly rate range | **Minimum rate ($/hour)** and **Desired rate ($/hour)** are required positive numbers; values are rounded to integers, desired must be at least minimum, and the UI caps both at $180/hour. Defaults delivered publicly are $20 and $25. The form says Lemon.io takes no commission from the stated amount. | [S1](https://me.lemon.io/escape-the-matrix) |
| 3 | Main specialization | A specialization and **Seniority level** are required. The public list currently contains 39 roles, from AI Agent Architect through WordPress Developer, and permits a custom specialization. Seniority options are Junior, Junior-to-middle, Middle, Middle-to-senior, Senior, and Strong senior. | [S1](https://me.lemon.io/escape-the-matrix) |
| 4 | Commercial experience | Required start year, defined by the form as when the applicant first received money for a project. Accepted years run from 1970 through the current year. | [S1](https://me.lemon.io/escape-the-matrix) |
| 5 | Core technologies | Select **1 to 6** present core skills and specify recent practical experience for each. A skill can be selected from the delivered catalog or entered as “Other”; experience choices run in half-year increments from 0.5 to 20 years. | [S1](https://me.lemon.io/escape-the-matrix) |
| 6 | Experience highlights | **LinkedIn**, **Link to CV**, and **Upload CV** controls are observable. The screen says “Add cv or linkedin … (at least one is required),” and advancement requires at least one of LinkedIn, CV link, or uploaded CV. | [S1](https://me.lemon.io/escape-the-matrix) |
| 6 | CV upload constraints | Accepted MIME types are PDF, Microsoft Word (`application/msword`), JPEG/JPG, and PNG; maximum size is 20 MiB. No `.docx` MIME type is present in the publicly delivered allow-list. | [S1](https://me.lemon.io/escape-the-matrix) |
| 7 | English readiness | One of two choices is required: written communication only, or ability to handle an interview and day-to-day work with English-speaking teammates. The form warns that actual English will be assessed in the first screening steps. | [S1](https://me.lemon.io/escape-the-matrix) |
| 8 | Registration details | Required: **First name**, **Last name**, valid **Email**, **Country of residence**, and agreement to Terms of Use/Privacy Policy. **Phone number** is explicitly optional. The action is labeled **Register** and says it creates an account and guides the applicant to next steps. | [S1](https://me.lemon.io/escape-the-matrix) |

### Current specialization options

The 39 values directly delivered by the live form are: AI Agent Architect; AI Automation Architect; AI Engineer; AI Red Teamer; Animator; Architect; Automation QA Engineer; Back-end Web Developer; Blockchain Engineer; CTO; Cloud Engineer; Data Analyst; Data Annotator; Data Engineer; Data Scientist; DevOps; DevSecOps; Embedded Software Engineer; Engineering Manager; Front-end Web Developer; Full-stack Web Developer; Game Developer; Graphic Designer; MLOps Engineer; Machine Learning Engineer; Manual QA Engineer; Mobile Developer; Platform Engineer; Product Designer; Product Manager; Project Manager; Prompt Engineer; Robotics Engineer; Shopify Developer; Site Reliability Engineer; Team lead; Tech lead; UI/UX Designer; WordPress Developer. [S1](https://me.lemon.io/escape-the-matrix)

The form also delivers a large technology catalog and specialization-specific suggested skills. Because these are selectable catalog values rather than eligibility requirements, they are not reproduced as “requirements.” [S1](https://me.lemon.io/escape-the-matrix)

## Published eligibility requirements

| Verified published fact | Primary source |
|---|---|
| Lemon.io describes itself as a senior-only network and says a likely fit has at least **3 years of commercial experience in a specific stack**; most network developers have 6-8 years. | [S3](https://docs.lemon.io/for-developers/who-can-apply.md) |
| Proven work on real products matters; tutorials and side projects do not satisfy the stated commercial-experience standard. | [S3](https://docs.lemon.io/for-developers/who-can-apply.md) |
| Strong spoken and written English sufficient for standups, asynchronous discussion, and technical conversations is required; clarity matters rather than accent or grammatical perfection. | [S3](https://docs.lemon.io/for-developers/who-can-apply.md) |
| Stated availability requirement is **20+ hours/week**; full-time (40 hours/week) is preferred and generally matches faster. | [S3](https://docs.lemon.io/for-developers/who-can-apply.md) |
| Lemon.io seeks relevant in-demand stacks across mainstream web/mobile/cloud technologies and Data, AI/ML, QA, and UX/UI roles. | [S3](https://docs.lemon.io/for-developers/who-can-apply.md) |
| Applicants on notice may apply, ideally with interim part-time availability; profiles needing 2+ months before full-time work may not advance immediately. | [S3](https://docs.lemon.io/for-developers/who-can-apply.md) |
| Junior and mid-level engineers may register, but Lemon.io says it will contact them once they reach client-required seniority. This explains why the live form offers non-senior values despite the published “senior-only” eligibility statement. | [S3](https://docs.lemon.io/for-developers/who-can-apply.md), [S1](https://me.lemon.io/escape-the-matrix) |

## Documents, checks, and verification

| Stage | Verified fact | Primary source |
|---|---|---|
| Initial form | At least one of LinkedIn, CV link, or CV upload is required by the live public form. | [S1](https://me.lemon.io/escape-the-matrix) |
| Profile review | Official help says Lemon.io reviews the resume, LinkedIn profile, and portfolio and performs an **identity and consistency check**. This is a published later review description, not evidence that all three are required fields in the current initial form. | [S4](https://docs.lemon.io/vetting-developers.md) |
| Background checks | The official help landing page says Lemon.io conducts multi-source background checks of engineers, but does not publicly specify the sources, documents, timing, or applicant actions. | [S9](https://docs.lemon.io/) |
| Identity documents | No passport, government-ID, selfie, proof-of-address, certificate, diploma, reference, or right-to-work upload is observable in the public eight-step form. Official pages reviewed here do not specify such an upload during application. This is a scoped negative observation, not proof that none is requested later. | [S1](https://me.lemon.io/escape-the-matrix), [S4](https://docs.lemon.io/vetting-developers.md) |
| Portfolio / GitHub | Portfolio is named as material reviewed during profile review, but no dedicated portfolio or GitHub field is observable in the initial public form. A URL could potentially be included in a CV, CV link, or LinkedIn profile, but that is an inference and is not stated as a requirement. | [S4](https://docs.lemon.io/vetting-developers.md), [S1](https://me.lemon.io/escape-the-matrix) |

## Later stages

Official sources describe overlapping stage models rather than one perfectly consistent sequence.

| Stage / behavior | Verified public description | Primary source |
|---|---|---|
| Profile creation and role choice | Upload CV and provide LinkedIn to create a profile, then select a role/stack that determines technical assessments. This marketing-page wording is stricter than the live form, which requires at least one of CV or LinkedIn. | [S2](https://lemon.io/for-developers/) and live-form contrast [S1](https://me.lemon.io/escape-the-matrix) |
| Automated or human pre-screening | Lemon.io's AI assistant confirms experience, availability, time zone, rates, and desired project types in an audio-only call advertised as 20 minutes or less. Applicants may choose a human interviewer instead, with access to the same projects. | [S2](https://lemon.io/for-developers/) |
| Technical pre-screen | The developer page describes a **15-minute role-specific task**. Official help describes role-specific multiple-choice and open-ended questions, with recruiting review and possible override of automated results. | [S2](https://lemon.io/for-developers/), [S4](https://docs.lemon.io/vetting-developers.md) |
| Recruiter stage | The developer page describes a focused **20-minute** conversation about work style and communication. Another official help page describes a **40-minute** live soft-skills/English interview covering English, communication, conduct, and CV verification. A third official page describes a live **20-minute** recruiter call after pre-screening. These durations conflict. | [S2](https://lemon.io/for-developers/), [S6](https://docs.lemon.io/for-developers/soft-skill-vetting.md), [S4](https://docs.lemon.io/vetting-developers.md) |
| Recruiter-stage conditionality | The current public vetting page says the recruiter interview can be skipped if earlier stages suffice. | [S8](https://lemon.io/our-vetting-process/) |
| Technical interview | A senior engineer/CTO conducts real-world problem solving including experience discussion, theory, system design/live coding/code review/bug fixing or algorithm analysis, plus AI-tool fluency. | [S4](https://docs.lemon.io/vetting-developers.md), [S8](https://lemon.io/our-vetting-process/) |
| Technical interview duration | Official developer help says approximately **90 minutes**. | [S5](https://docs.lemon.io/for-developers/tech-vetting-criteria.md) |
| Technical decision | Technical-interview feedback is shared within one business day according to technical help; the marketing page instead says applicants learn whether they made the cut “a few days later.” Inconclusive feedback gets a second specialist on the same stack. | [S5](https://docs.lemon.io/for-developers/tech-vetting-criteria.md), [S2](https://lemon.io/for-developers/) |
| Acceptance | Lemon.io publishes a **1.2%** overall acceptance rate; the current vetting page rounds the listed-on-platform outcome to roughly 1%. | [S9](https://docs.lemon.io/), [S8](https://lemon.io/our-vetting-process/) |
| Onboarding | After passing vetting, validated skills/stacks are finalized; only technically verified skills are shown to clients, and the full profile is not public. | [S7](https://docs.lemon.io/for-developers/developer-onboarding.md) |
| Matching and projects | Offers are matched by stack, rate, and availability and sent through a personal Slack channel. Developers can also browse open platform projects, but do not bid or write cover letters. | [S7](https://docs.lemon.io/for-developers/developer-onboarding.md) |
| First offer | Official onboarding help says the first offer arrives in about **13 days on average after vetting**, depending on stack demand, rate, and availability. This is not application-to-approval time. | [S7](https://docs.lemon.io/for-developers/developer-onboarding.md) |

## Timing evidence and conflicts

- The developer landing page promises “application to approval in days” but gives no fixed total. It says the final decision follows a few days after the technical interview. [S2](https://lemon.io/for-developers/)
- The soft-skill help page says the vetting process spans **approximately two weeks** and involves multiple touchpoints. [S6](https://docs.lemon.io/for-developers/soft-skill-vetting.md)
- Technical-interview feedback is promised within **one business day** by technical help. [S5](https://docs.lemon.io/for-developers/tech-vetting-criteria.md)
- The **~13 days** figure is average time from passing vetting to a first project offer, not time to approve an application. [S7](https://docs.lemon.io/for-developers/developer-onboarding.md)
- Therefore, no current primary source reviewed verifies an end-to-end **~5-day** application process. The official timing claims are inconsistent enough that a single replacement duration should not be inferred.

## Auth-only and inaccessible portions

- The initial eight-step flow is public, but its last action registers an account; it was not invoked. [S1](https://me.lemon.io/escape-the-matrix)
- `/my-application` redirects unauthenticated requests to Lemon.io's login page. Consequently, post-registration questionnaires, scheduling interfaces, assessment questions, application status controls, account/profile fields, and any later upload or verification prompts were not directly inspected. [S10](https://me.lemon.io/my-application), [S11](https://app.lemon.io/auth/login)
- Public source bundles name internal statuses and routes, but those implementation strings are not evidence that a particular applicant will see a field or undergo a step; they are intentionally excluded as requirements.

## Legacy-summary comparison

| Legacy item | Current determination | Evidence |
|---|---|---|
| CV | **Verified, with nuance.** Current form requires at least one of CV (link/upload) or LinkedIn, not necessarily CV alone. Marketing copy says upload CV and provide LinkedIn; the live form is the strongest direct field evidence. | [S1](https://me.lemon.io/escape-the-matrix), [S2](https://lemon.io/for-developers/) |
| Stack | **Verified.** Main specialization, seniority, commercial-experience year, and 1-6 technologies with years are public form fields; a relevant stack is a published eligibility criterion. | [S1](https://me.lemon.io/escape-the-matrix), [S3](https://docs.lemon.io/for-developers/who-can-apply.md) |
| Target rate | **Verified and expanded.** The form requests both minimum and desired hourly rates. | [S1](https://me.lemon.io/escape-the-matrix) |
| Availability | **Verified.** Initial project type captures full-time, part-time, or unavailable/looking around; published eligibility is 20+ hours/week. | [S1](https://me.lemon.io/escape-the-matrix), [S3](https://docs.lemon.io/for-developers/who-can-apply.md) |
| Time zone | **Verified as a later pre-screen topic, not an observable initial-form field.** | [S2](https://lemon.io/for-developers/), [S1](https://me.lemon.io/escape-the-matrix) |
| Start date | **Not directly verified as a current field.** Notice-period/start readiness affects progression, but no explicit start-date field is observable in the public form or stated in the reviewed process pages. | [S3](https://docs.lemon.io/for-developers/who-can-apply.md), [S1](https://me.lemon.io/escape-the-matrix) |
| Soft-skills and English interview | **Verified.** Duration is inconsistent across current primary pages: 20 versus 40 minutes. | [S2](https://lemon.io/for-developers/), [S4](https://docs.lemon.io/vetting-developers.md), [S6](https://docs.lemon.io/for-developers/soft-skill-vetting.md) |
| Technical interview | **Verified.** It follows pre-screening/recruiter stages and is approximately 90 minutes according to developer help. | [S4](https://docs.lemon.io/vetting-developers.md), [S5](https://docs.lemon.io/for-developers/tech-vetting-criteria.md) |
| Optional GitHub/portfolio | **Partially verified.** Portfolio is reviewed according to official help, but neither portfolio nor GitHub has a dedicated observable public-form field. Optionality is not explicitly stated. | [S4](https://docs.lemon.io/vetting-developers.md), [S1](https://me.lemon.io/escape-the-matrix) |
| ~5 days | **Not verified/currently unsupported.** Current sources say “in days,” “a few days later,” approximately two weeks for vetting, and ~13 days after vetting for a first offer. | [S2](https://lemon.io/for-developers/), [S6](https://docs.lemon.io/for-developers/soft-skill-vetting.md), [S7](https://docs.lemon.io/for-developers/developer-onboarding.md) |

## Inferences and limits

1. **Inference:** Because the public form ends with “Register” and `/my-application` is login-gated, additional workflow exists after the eight public screens. Its exact fields cannot be verified without prohibited account creation/authentication. [S1](https://me.lemon.io/escape-the-matrix), [S10](https://me.lemon.io/my-application)
2. **Inference:** The live form's “CV or LinkedIn” validation is likely the operative initial-document rule, while marketing/help descriptions of CV plus LinkedIn/portfolio describe preferred inputs or review scope. This cannot establish whether recruiters later demand omitted material. [S1](https://me.lemon.io/escape-the-matrix), [S2](https://lemon.io/for-developers/), [S4](https://docs.lemon.io/vetting-developers.md)
3. **Unknown:** Exact conditional branching, qualification thresholds, assessment content, scheduling availability, rejection/reapplication UI, identity-check mechanics, and later uploads are not publicly observable in this read-only unauthenticated review.
