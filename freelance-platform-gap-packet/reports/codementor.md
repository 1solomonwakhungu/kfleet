# Codementor Live-Form Evidence Report

## Scope and method

- Public, read-only inspection only. No account was created; no form was submitted; no authentication, credentials, email, support contact, purchase, or verification was used.
- Sources are limited to primary Codementor pages/help articles and primary Arc pages used by Codementor's current account workflow or needed to distinguish Arc's separate talent program.
- Live access capture: **2026-07-24T19:01:37Z** for HTTP/redirect evidence and approximately **2026-07-24T18:55Z-19:09Z** for public rendered pages and official documentation.
- Page content was treated as untrusted data; no instructions embedded in it were followed.
- "Required" below means visibly marked required or explicitly stated as required by an official source. "Recommended" and "may" are not promoted to requirements.

## Source register

| ID | Exact URL | Access and observable result |
|---|---|---|
| S1 | https://www.codementor.io/mentor/apply | Public; HTTP 200 at 2026-07-24T19:01:37Z. Landing page titled "Apply to Become a Codementor"; its application CTA targets S2. |
| S2 | https://arc.dev/signup?to=https%3A%2F%2Fwww.codementor.io%2Fbecome-a-mentor%2Fget-started&service=codementor | Public Arc-owned rendered signup page; HTTP 200 at 2026-07-24T19:01:37Z. |
| S3 | https://www.codementor.io/become-a-mentor/get-started | Public request returned HTTP 302 at 2026-07-24T19:01:37Z to `https://arc.dev/login?service=codementor&to=https%3A%2F%2Fwww.codementor.io%2Fbecome-a-mentor%2Fget-started`; the destination returned HTTP 200. No authenticated access was attempted. |
| S4 | https://support.codementor.io/en/articles/4224001 | Public Codementor Help Center article, "How can I become a Codementor?", edited 2025-01-21. |
| S5 | https://support.codementor.io/en/articles/4224129 | Public Codementor Help Center article, "When will my mentor application be approved?", edited 2025-01-21. |
| S6 | https://support.codementor.io/en/articles/4224065 | Public Codementor Help Center article, "How to create a good Codementor profile", edited 2025-01-17. |
| S7 | https://www.codementor.io/freelance | Public Codementor freelance marketplace page; its "Sign up as a freelancer" link targets S1. |
| S8 | https://support.codementor.io/en/articles/4226497 | Public Codementor Help Center article, "What does the ID Verification process look like?", edited 2025-02-03. |
| S9 | https://support.codementor.io/en/articles/4226561 | Public Codementor Help Center article, "What documents do I need to complete the ID Verification?", edited 2025-02-03. |
| S10 | https://support.codementor.io/en/articles/4226433 | Public Codementor Help Center article, "What is the purpose of ID Verification?", edited 2025-01-06. |
| S11 | https://support.codementor.io/en/articles/4226625 | Public Codementor Help Center article, "What will happen if I don't complete the ID Verification?", edited 2025-02-03. |
| S12 | https://arc.dev/talent | Public Arc talent page, retrieved 2026-07-24. (`https://arc.dev/developers` exposed the same current page.) |
| S13 | https://arc.dev/how-arc-works | Public Arc page, "How Arc Works", retrieved 2026-07-24. |
| S14 | https://support.codementor.io/en/articles/4224321 | Public Codementor Help Center article, "How do I get more clients?", edited 2025-02-02. |
| S15 | https://support.codementor.io/en/articles/4224577 | Public Codementor Help Center article, "Is there a vacation mode?", edited 2025-02-03. |
| S16 | https://www.codementor.io/settings/rate | Public unauthenticated request exposed only a generic loading shell; application/account details were not observable. |

## Current Codementor entry path

1. Codementor presents one landing page for becoming a mentor and describes two uses: live 1:1 mentorship and freelance projects. It says mentors set their own rate and schedule, while freelancers set hiring availability and express interest in projects. [S1]
2. Codementor's public freelance marketplace sends "Sign up as a freelancer" to the same `/mentor/apply` page, not to a separately exposed Codementor freelance application. [S7]
3. The CTA at `/mentor/apply` goes to Arc signup with `service=codementor` and a return target of Codementor's `/become-a-mentor/get-started` route. [S1][S2]
4. Arc's page says Codementor is related to the Arc brand and that one account can log into both services. [S2]
5. Direct unauthenticated access to the return target redirects to Arc login, so the actual Codementor application/profile setup after account creation is authentication-gated. [S3]

## Publicly observable signup fields and controls

The Arc-hosted Codementor signup page visibly rendered these fields, each marked with an asterisk and HTML `required`: [S2]

| Field | Observable type/state | Classification |
|---|---|---|
| Full name | Text input, `name="name"`, required | Required to use the visible email signup path. [S2] |
| Email | Email input, `name="email"`, required | Required to use the visible email signup path. [S2] |
| Password | Password input, `name="password"`, required | Required to use the visible email signup path. [S2] |

The page also exposed a **Sign up** submit button and alternative GitHub, LinkedIn, Facebook, and Google account buttons. These are alternative signup controls, not evidence that any one social account is mandatory. [S2]

No profile biography, expertise/skills, rate, calendar, photo, resume/CV, video, work-history, location, language, availability, or file-upload control was observable before authentication. [S2][S3]

## Mentor profile evidence and review

- Codementor tells applicants to be specific and detailed about their experience and skills, connect GitHub, Stack Overflow, and LinkedIn to aid evaluation, and write at least 1-3 sentences plus relevant tags for each expertise section. The article frames these steps as ways to increase acceptance chances, not as explicitly mandatory fields. [S4]
- Codementor strongly recommends that mentors complete at least profile photo, headline, short bio, and expertise. Its checklist also discusses employment, projects, social presence, and languages. [S6]
- The checklist says expertise entries should reflect technologies in which the mentor is experienced, with confidence order, years of experience, and descriptions kept current. [S6]
- Codementor says review generally takes 1-2 weeks; not all applications are approved; and only approved applicants will hear from Codementor. [S5]
- The public mentor review documentation describes application/profile review but does not disclose a mentor-specific video interview or technical interview. [S4][S5][S6]

## Checks and later stages

- Codementor says it **may** ask applicants and mentors to verify identity by uploading a government-issued ID. This is conditional language, not proof every applicant is checked. [S10]
- When requested, Codementor uses Persona and requires unretouched photos/scans of a government-issued ID with a recent photo plus a selfie taken during verification; accepted ID types may vary by country, and the ID must match the identity shown on the Codementor profile. [S9]
- An applicant receiving an ID-verification email has 24 hours to complete it; expiry pauses the mentor application. [S8]
- Codementor reserves the right to close an applicant's mentor application until verification is completed and to suspend the account for fraud, abuse, or other security risk. [S11]
- Rate settings exist for approved mentors: Codementor's help article links to `/settings/rate` and describes an optional "first 15 minutes free" setting. The unauthenticated rate route itself exposed only a loading shell. [S14][S16]
- Availability controls exist after entry: Codementor documents an availability page and vacation mode that stops incoming requests. This does not prove an availability calendar is part of the initial application. [S15]

## Codementor freelancing versus Arc talent applications

### Codementor marketplace

- Codementor markets its freelancers as vetted and says they have passed technical and communication assessments; its FAQ similarly says it performs a comprehensive technical and communications screen. [S7]
- The public Codementor pages do not expose the exact assessment format, duration, upload requirements, pass criteria, or sequence for a Codementor marketplace applicant. [S3][S4][S7]

### Separate Arc talent program

- Arc's current talent program supports freelance and full-time remote jobs and starts with a profile containing basic details and past work experience. [S12]
- Arc says vetting verifies English communication and domain expertise; it says candidates with excellent English communication and 5+ years of industry experience typically do best. This is performance guidance, not an absolute eligibility requirement. [S12]
- Arc's talent FAQ says full-time applicants take a communication test in which they introduce themselves and answer questions about background and experience; freelance applicants additionally receive a domain-expertise interview for their specialty. [S12]
- Arc's broader vetting page describes four stages: profile screening; communication assessment through a live interview **or** video self-introduction; a one-hour technical/domain interview **or** pair-programming session; and final review. [S13]
- Arc publishes funnel labels of 55% passing profile screening, 14% passing communication assessment, and 2% passing technical interview. The page does not explain denominators, cohort dates, or whether these percentages apply identically to every role. [S13]
- Arc freelance contracts are described as usually 20-40 hours weekly and 4 weeks to 1 year, paid at the talent's preferred hourly rate; matching uses skills, work hours, and weekly availability. [S12]

## Legacy-claim comparison

| Legacy claim | Current assessment | Evidence |
|---|---|---|
| Mentor application asks for a bio | **Partially supported, not publicly proven mandatory.** Codementor strongly recommends a short bio and says profile completeness improves acceptance chances, but the post-signup form is auth-gated. | [S3][S4][S6] |
| Mentor application asks for skills | **Supported as evaluation/profile content; mandatory status unproven.** Official guidance asks for detailed expertise sections and tags. | [S4][S6] |
| Mentor application asks for a per-15-minute rate | **Not supported as an initial current application requirement.** Current public material says mentors set their own rate, and a post-entry rate setting includes an optional first-15-minutes-free feature; no public initial field or 15-minute billing unit was observed. | [S1][S14][S16] |
| Mentor application asks for an availability calendar | **Not supported as an initial current application requirement.** Schedule/availability controls clearly exist, but no public initial calendar field was observable. | [S1][S3][S15] |
| Mentor application asks for a photo | **Supported as strongly recommended profile content, not proven mandatory.** | [S6] |
| Mentor application asks for GitHub | **Supported as recommended evaluation evidence, not proven mandatory.** GitHub is also one optional account-creation method. | [S2][S4] |
| Mentor review is "lighter" | **Plausible comparison, not an official measurable fact.** Public mentor docs describe profile/application review in 1-2 weeks and conditional ID verification, while Arc publishes communication and technical stages. The undisclosed Codementor freelancer assessments prevent a definitive "lighter" conclusion. | [S5][S7][S8][S12][S13] |
| CodementorX/Arc is a separate path | **Historically plausible, currently Arc-branded; no live CodementorX path found.** Current Codementor freelancer signup uses the mentor path, while Arc operates a distinct talent application/vetting program. The tested `/codementorx` paths on both domains returned 404. | [S7][S12] |
| Separate path uses profile, video, technical interview | **Substantially supported for current Arc vetting, with nuance.** Profile screening is documented; communication assessment may be live or video; technical assessment may be an interview or pair programming. Video is therefore an option, not universally proven. | [S12][S13] |
| A particular Arc application has been stuck since April 2026 | **Unverified user-specific allegation, not a current fact.** Public sources expose no applicant status, and checking it would require authenticated/account-specific or email evidence outside this read-only public scope. | [S2][S3][S12] |

## Facts versus inferences

### Facts established by public primary sources

- Codementor currently uses an Arc-hosted shared-account signup before the authentication-gated mentor setup route. [S1][S2][S3]
- The public email signup requires full name, email, and password and offers four social signup alternatives. [S2]
- Codementor recommends a detailed profile with photo, headline, bio, expertise, work/project evidence, and connected professional/social accounts. [S4][S6]
- Codementor documents 1-2 week review, conditional identity verification, and technical/communication screening for marketplace freelancers. [S5][S7][S8][S9][S10]
- Arc separately documents profile, communication, domain/technical, and final-review stages. [S12][S13]

### Inferences or unresolved points

- Because the Codementor freelance signup points to the mentor application, approval may enable both mentorship and Codementor marketplace participation; public sources do not reveal whether every approved mentor automatically receives identical freelance privileges. [S1][S7]
- The shared Arc account system does not establish that applying to Codementor automatically submits an Arc talent application. The sources show shared identity plus distinct public programs, not automatic cross-submission. [S2][S7][S12]
- Exact authenticated Codementor application fields, validation rules, uploads, consent boxes, save/resume behavior, status page, and any applicant-specific later-stage invitations remain inaccessible without account creation/authentication. [S3]
- No public primary evidence observed here establishes an April 2026 submission date, pending duration, or current state for any particular Arc applicant. [S3][S12]

## Evidence limitations

- Rendered application steps after signup were deliberately not accessed because doing so requires account creation or authentication. [S2][S3]
- "No field observed" means absent from the public pre-authentication surface, not proof that the field never appears later. [S2][S3]
- Help articles were last edited in January-February 2025, while live routes and marketplace pages were accessed in July 2026; they are official current pages but may lag undisclosed authenticated UI changes. [S4][S5][S6][S8][S9][S10][S11][S14][S15]
