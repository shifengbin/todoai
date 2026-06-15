## MODIFIED Requirements

### Requirement: Classify Terminal Activity From Titles

The system SHALL classify terminal activity from captured terminal title changes without treating ordinary Windows path separators, stable program titles, or single-frame Claude title markers as busy indicators. Title-derived classification SHALL be treated as a low-confidence fallback and MUST NOT override a newer higher-priority unified agent status from shell lifecycle, command-state, Claude/Codex structured events, or machine-readable agent streams.

#### Scenario: Windows path title remains idle

- **WHEN** the application runs on Windows
- **AND** terminal A receives a title update `C:\Users\developer\repo`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's title fallback activity state remains `idle`
- **AND** the TODO terminal tree does not display terminal A as busy solely because the title contains `\`

#### Scenario: Unix path title remains idle

- **WHEN** terminal A receives a title update `/home/developer/repo`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's title fallback activity state remains `idle`
- **AND** the TODO terminal tree does not display terminal A as busy solely because the title contains `/`

#### Scenario: Explicit busy title signal marks terminal busy

- **WHEN** terminal A receives a title update `codex thinking`
- **AND** terminal A has no newer structured agent status
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's unified agent activity phase becomes `busy`
- **AND** the status source is `title-fallback`

#### Scenario: Spinner title signal marks terminal busy

- **WHEN** terminal A receives a title update containing an explicit spinner character such as `⠋`
- **AND** terminal A has no newer structured agent status
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's unified agent activity phase becomes `busy`
- **AND** the status source is `title-fallback`

#### Scenario: Needs input title signal marks terminal needs input

- **WHEN** terminal A receives a title update `codex !`
- **AND** terminal A has no newer structured agent status
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's unified agent activity phase becomes `needs-input`
- **AND** the status source is `title-fallback`

#### Scenario: Stable title establishes idle baseline

- **WHEN** terminal A is running without a prior idle title
- **AND** terminal A receives a stable title update matching its shell or current command label
- **THEN** the system records that title as terminal A's idle title
- **AND** terminal A's title fallback activity state remains `idle`

#### Scenario: Structured status is not overridden by title fallback

- **WHEN** terminal A has unified agent activity phase `needs-input` from a Claude hook notification
- **AND** terminal A receives a terminal title update `claude thinking`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's unified agent activity phase remains `needs-input`
- **AND** terminal A's unified agent status source remains the Claude hook source
