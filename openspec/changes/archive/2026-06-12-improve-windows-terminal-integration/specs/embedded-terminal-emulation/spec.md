## ADDED Requirements

### Requirement: Classify Terminal Activity From Titles

The system SHALL classify terminal activity from captured terminal title changes without treating ordinary Windows path separators as busy indicators. The system SHALL continue to surface explicit busy and needs-input title signals.

#### Scenario: Windows path title remains idle

- **WHEN** the application runs on Windows
- **AND** terminal A receives a title update `C:\Users\developer\repo`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state remains `idle`
- **AND** the TODO terminal tree does not display terminal A as busy solely because the title contains `\`

#### Scenario: Unix path title remains idle

- **WHEN** terminal A receives a title update `/home/developer/repo`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state remains `idle`
- **AND** the TODO terminal tree does not display terminal A as busy solely because the title contains `/`

#### Scenario: Explicit busy title signal marks terminal busy

- **WHEN** terminal A receives a title update `codex thinking`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state becomes `busy`

#### Scenario: Spinner title signal marks terminal busy

- **WHEN** terminal A receives a title update containing an explicit spinner character such as `⠋`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state becomes `busy`

#### Scenario: Needs input title signal marks terminal needs input

- **WHEN** terminal A receives a title update `codex !`
- **THEN** the system records terminal A's latest runtime title
- **AND** terminal A's activity state becomes `needs-input`

#### Scenario: Stable title establishes idle baseline

- **WHEN** terminal A is running without a prior idle title
- **AND** terminal A receives a stable title update matching its shell or current command label
- **THEN** the system records that title as terminal A's idle title
- **AND** terminal A's activity state remains `idle`
