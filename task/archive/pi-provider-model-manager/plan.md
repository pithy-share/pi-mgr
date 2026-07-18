# Implementation Plan: Pi Provider & Model Manager

## Product Invariants

From PRD + pi docs analysis:

1. **models.json schema**: Top-level `{"providers": {}}` object. Each provider keyed by provider ID (string). Provider fields: `baseUrl`, `api`, `apiKey`, `models[]`, `headers`, `authHeader`, `oauth`, `modelOverrides`, `compat`. Model fields: `id` (required), `name`, `api`, `reasoning`, `input`, `contextWindow`, `maxTokens`, `cost`, `compat`.

2. **Built-in provider override (no models)**: When only `baseUrl` and/or `apiKey` are set without `models[]`, pi preserves all built-in models for that provider, just routes through the new endpoint.

3. **Built-in provider with models**: Custom models upsert by `id` — matching built-in IDs are replaced, new IDs are added alongside built-ins. `api` and `compat` are not required since pi already knows them for built-in providers.

4. **Custom provider (non-built-in)**: Must have `baseUrl` and `api` at provider or model level. Must have `models[]` to define available models. `apiKey` is optional (can be provided via `/login`, `auth.json`, or CLI).

5. **One-to-one**: Each scheme is independent; only one scheme is "active" (written to models.json) at a time. A scheme contains a set of providers, each with a set of models. A built-in provider appears at most once per scheme (AC-33). Custom provider names must be unique per scheme (AC-29). Model IDs must be unique per provider per scheme (AC-28).

6. **No network validation**: Tool never validates API keys or endpoints. No HTTP calls.

7. **Windows-only**: `~/.pi/agent/models.json` resolves to `%USERPROFILE%/.pi/agent/models.json`.

## Implementation Steps

### Phase 1: Wails Project Scaffold & Data Model (Go backend)

**Step 1.1: Initialize Wails project**
- Target: `d:/src/pi-mgr/`
- Command: `wails init -n pi-mgr -t vue-ts` (or vanilla if preferred)
- This creates the Go backend at root level and frontend in `frontend/`
- Validation: `wails dev` starts the app successfully with default template

**Step 1.2: Define Go data structures**
- File: `models.go` (new)
- Define types mirroring the scheme storage model (NOT the models.json output format):

```go
// Scheme is a named collection of providers
type Scheme struct {
    ID        string     `json:"id"`
    Name      string     `json:"name"`
    Providers []Provider `json:"providers"`
}

// Provider represents either a built-in or custom provider config
type Provider struct {
    Key        string  `json:"key"`         // provider ID (e.g., "openai", "my-custom")
    Name       string  `json:"name"`        // display name for custom; empty for built-in
    BuiltIn    bool    `json:"builtIn"`     // true = built-in, false = custom
    APIKey     string  `json:"apiKey"`      // literal API key string
    BaseURL    string  `json:"baseUrl"`     // optional override
    APIType    string  `json:"apiType"`     // for custom: openai-completions / anthropic-messages / google-generative-ai
    Models     []Model `json:"models"`
}

type Model struct {
    ID            string  `json:"id"`
    Name          string  `json:"name"`
    Reasoning     bool    `json:"reasoning"`
    InputText     bool    `json:"inputText"`
    InputImage    bool    `json:"inputImage"`
    ContextWindow int     `json:"contextWindow"`
    MaxTokens     int     `json:"maxTokens"`
    CostInput     float64 `json:"costInput"`
    CostOutput    float64 `json:"costOutput"`
    CostCacheRead float64 `json:"costCacheRead"`
    CostCacheWrite float64 `json:"costCacheWrite"`
}

// Built-in provider catalog entry
type BuiltInProvider struct {
    Key     string `json:"key"`
    Name    string `json:"name"`
    APIType string `json:"apiType"`
}
```

- Invariants maintained: Each Provider.Key is unique per Scheme. Each Model.ID is unique per Provider.
- Verification: Compiles; types match all PRD AC field requirements.

**Step 1.3: Implement storage layer**
- File: `store.go` (new)
- Use a JSON file at `%APPDATA%/pi-mgr/schemes.json` for persistence (AC-22, AC-23)
- Functions:
  - `LoadSchemes() ([]Scheme, error)` — read from disk; return empty slice if file doesn't exist
  - `SaveSchemes([]Scheme) error` — atomic write (write to temp + rename)
  - `GetScheme(id string) (*Scheme, error)`
  - `CreateScheme(name string) (*Scheme, error)` — generate UUID for ID
  - `UpdateScheme(scheme Scheme) error`
  - `DeleteScheme(id string) error`
  - `DuplicateScheme(id string) (*Scheme, error)` — new UUID, name += " - 副本"
- Concurrency: No multi-instance detection (out of scope). Single process, single writer.
- Verification: AC-01 (list on startup), AC-02~05 (CRUD + duplicate), AC-22 (persist across restart), AC-23 (decoupled from models.json), AC-30 (cancel = no delete).

**Step 1.4: Define built-in provider catalog**
- File: `builtin.go` (new)
- Hardcoded list per AC-08:

```go
var BuiltInProviders = []BuiltInProvider{
    {Key: "openai",       Name: "OpenAI",           APIType: "openai-completions"},
    {Key: "anthropic",    Name: "Anthropic",         APIType: "anthropic-messages"},
    {Key: "deepseek",     Name: "DeepSeek",          APIType: "openai-completions"},
    {Key: "google",       Name: "Google Gemini",     APIType: "google-generative-ai"},
    {Key: "mistral",      Name: "Mistral",           APIType: "mistral-conversations"},
    {Key: "groq",         Name: "Groq",              APIType: "openai-completions"},
    {Key: "xai",          Name: "xAI",               APIType: "openai-completions"},
    {Key: "openrouter",   Name: "OpenRouter",        APIType: "openai-completions"},
    {Key: "together",     Name: "Together AI",       APIType: "openai-completions"},
    {Key: "fireworks",    Name: "Fireworks",         APIType: "openai-completions"},
    {Key: "cerebras",     Name: "Cerebras",          APIType: "openai-completions"},
    {Key: "bedrock",      Name: "Amazon Bedrock",    APIType: "bedrock-converse-stream"},
    {Key: "nvidia",       Name: "NVIDIA NIM",        APIType: "openai-completions"},
    {Key: "huggingface",  Name: "Hugging Face",      APIType: "openai-completions"},
}
```

- Also define a `validAPITypes` list for custom provider dropdown: `openai-completions`, `anthropic-messages`, `google-generative-ai`
- Verification: AC-08 (list available), AC-13 (custom API type options).

**Step 1.5: Implement models.json serializer**
- File: `serializer.go` (new)
- Function: `SerializeToModelsJSON(scheme *Scheme) ([]byte, error)`
- Logic:
  1. Build `map[string]interface{}` for the `providers` object
  2. For each provider in the scheme:
     - If `BuiltIn`:
       - If `BaseURL == "" && APIKey == "" && len(Models) == 0` → **skip** (AC-20)
       - Else: create provider entry with only non-empty fields: `apiKey` (if set), `baseUrl` (if set), `models` (if any) (AC-12)
       - Do NOT emit `api` field for built-in providers
     - If `!BuiltIn` (custom):
       - Must have `baseUrl` and `apiType` (validated before save). Emit: `baseUrl`, `api`, `apiKey` (if non-empty), `models[]` (AC-15)
  3. For each model in provider.Models:
     - Always emit `id`
     - Emit `name` only if different from `id`
     - Emit `reasoning` only if true (omit `false` to keep JSON clean)
     - Build `input` array from InputText/InputImage booleans; omit if `["text"]` (default)
     - Emit `contextWindow` only if != 128000
     - Emit `maxTokens` only if != 16384
     - Emit `cost` only if any cost field != 0; omit zero-valued fields within cost
  4. Marshal to indented JSON
- Verification: AC-12, AC-15, AC-20, AC-21 (serialization correctness).

**Step 1.6: Implement models.json writer (activate)**
- File: `activate.go` (new)
- Function: `ActivateScheme(scheme *Scheme) error`
- Logic:
  1. Call `SerializeToModelsJSON(scheme)`
  2. Resolve `~/.pi/agent/` to `%USERPROFILE%/.pi/agent/`
  3. If directory doesn't exist, create it (AC-31)
  4. Write to `%USERPROFILE%/.pi/agent/models.json` (overwrite)
  5. Return error if write fails (AC-32)
- Verification: AC-06 (write on activate), AC-07 (success notification from frontend), AC-31, AC-32.

**Step 1.7: Implement validation logic**
- File: `validate.go` (new)
- Functions returning error messages:
  - `ValidateScheme(scheme *Scheme) []string` — name non-empty (AC-24)
  - `ValidateProvider(prov *Provider, allProviders []Provider) []string`:
    - Custom: baseUrl non-empty (AC-25), apiType selected (AC-26)
    - Custom: key not duplicate with existing (AC-29)
    - Built-in: not already in scheme (AC-33)
  - `ValidateModel(m *Model, existingModels []Model) []string`:
    - ID non-empty (AC-27)
    - ID not duplicate in same provider (AC-28)
- Called from frontend before save; backend also re-validates on API call
- Verification: AC-24 through AC-29, AC-33.

### Phase 2: Wails API Bindings (Go → Frontend bridge)

**Step 2.1: Define Wails bound API methods**
- File: `app.go` (modify existing or new)
- All methods use Wails runtime binding. Exported methods:

```go
// Scheme CRUD
func (a *App) ListSchemes() []Scheme
func (a *App) CreateScheme(name string) (*Scheme, error)
func (a *App) UpdateScheme(scheme Scheme) error
func (a *App) DeleteScheme(id string) error
func (a *App) DuplicateScheme(id string) (*Scheme, error)
func (a *App) ActivateScheme(id string) error

// Provider management (within a scheme)
func (a *App) AddBuiltInProvider(schemeID string, providerKey string, apiKey string, baseUrl string) error
func (a *App) AddCustomProvider(schemeID string, key string, baseUrl string, apiType string, apiKey string) error
func (a *App) UpdateProvider(schemeID string, provider Provider) error
func (a *App) RemoveProvider(schemeID string, providerKey string) error

// Model management (within a provider in a scheme)
func (a *App) AddModel(schemeID string, providerKey string, model Model) error
func (a *App) UpdateModel(schemeID string, providerKey string, model Model) error
func (a *App) RemoveModel(schemeID string, providerKey string, modelID string) error

// Catalogs
func (a *App) ListBuiltInProviders() []BuiltInProvider
func (a *App) ListAPITypes() []string
```

- Each mutating method: load schemes, find target, apply change, validate, save.
- Validation errors returned as Go `error` with user-facing message.
- Invariants: All state changes go through these methods. Frontend never touches disk directly.
- Verification: All ACs exercised through API; frontend integration tests can call these.

### Phase 3: Frontend (Vue + TypeScript)

**Step 3.1: Scheme list page**
- Route: `/` (home)
- Display list of schemes with name, edit/delete/duplicate/activate buttons (AC-01)
- "New scheme" button → inline name input or modal (AC-02)
- Edit name: inline edit or modal (AC-03)
- Duplicate: calls DuplicateScheme API, refreshes list (AC-04)
- Delete: confirmation dialog → calls DeleteScheme API (AC-05, AC-30)
- Activate: calls ActivateScheme API → toast on success/error (AC-06, AC-07)
- Empty state when no schemes (AC-01)

**Step 3.2: Scheme editor page**
- Route: `/scheme/:id`
- Left panel: Provider list (built-in with "(内置)" badge + custom with "(自定义)" badge) (AC-11, AC-14)
- Right panel: Provider detail + model list when a provider is selected
- "Add built-in provider" button → dropdown of available built-in providers not yet in scheme (AC-08, AC-33)
- "Add custom provider" button → form with key, baseUrl, apiType dropdown, apiKey (AC-13)

**Step 3.3: Provider config form**
- For built-in provider: API key field (password toggle) + optional baseUrl field (AC-09, AC-10, AC-12-1)
- For custom provider: key (read-only after create), baseUrl (required), apiType (dropdown, required), apiKey (optional, password toggle) (AC-13)
- Save button → calls UpdateProvider API
- Validation errors displayed inline (AC-24~29)
- Remove provider button (with confirmation)

**Step 3.4: Model list and editor**
- Within selected provider, show model list (AC-16)
- "Add model" button → form with fields per AC-17:
  - id (text, required) (AC-27)
  - name (text, defaults to id)
  - reasoning (checkbox, default false)
  - input types (checkboxes: text [always checked], image)
  - contextWindow (number, default 128000)
  - maxTokens (number, default 16384)
  - cost: input, output, cacheRead, cacheWrite (number inputs, default 0)
- Edit model: same form, pre-filled (AC-18)
- Delete model: confirmation dialog (AC-19, AC-30)
- Duplicate model id validation (AC-28)

### Phase 4: Integration & Polish

**Step 4.1: End-to-end flow validation**
- Create scheme → add built-in provider with API key → add custom provider with models → activate → verify models.json output
- Verify each AC manually

**Step 4.2: Build & package**
- `wails build` for Windows production binary
- Test on clean Windows environment

## Verification & Acceptance

| AC | Verification Method | Phase/Step |
|----|-------------------|------------|
| AC-01 | Manual: launch app, observe scheme list or empty state | 3.1 |
| AC-02 | Manual: create scheme, verify it appears | 3.1 |
| AC-03 | Manual: edit scheme name inline | 3.1 |
| AC-04 | Manual: duplicate scheme, verify " - 副本" suffix | 3.1 |
| AC-05 | Manual: delete with confirm/cancel | 3.1 |
| AC-06 | Manual: activate, check `%USERPROFILE%/.pi/agent/models.json` | 3.1 |
| AC-07 | Manual: activate, observe toast message | 3.1 |
| AC-08 | Manual: open "add built-in" dropdown, verify list and API types | 3.2 |
| AC-09 | Manual: select built-in, verify form shows apiKey + optional baseUrl | 3.3 |
| AC-10 | Manual: toggle password visibility on apiKey field | 3.3 |
| AC-11 | Manual: save, verify provider appears in list | 3.2 |
| AC-12 | Unit: serializer output for built-in with only apiKey has no models/api | 1.5 |
| AC-12-1 | Manual: baseUrl field present on built-in form | 3.3 |
| AC-13 | Manual: create custom provider form with all fields | 3.2, 3.3 |
| AC-14 | Manual: custom provider shows "自定义" badge in list | 3.2 |
| AC-15 | Unit: serializer output for custom has baseUrl+api+models | 1.5 |
| AC-16 | Manual: click provider, see model list | 3.4 |
| AC-17 | Manual: add model form with all fields and defaults | 3.4 |
| AC-18 | Manual: edit model, verify changes persist | 3.4 |
| AC-19 | Manual: delete model with confirm/cancel | 3.4 |
| AC-20 | Unit: serializer skips built-in with no apiKey+baseUrl+models | 1.5 |
| AC-21 | Unit: serializer includes built-in with models, custom models in array | 1.5 |
| AC-22 | Manual: create scheme, restart app, verify it persists | 1.3 |
| AC-23 | Manual: edit scheme, verify models.json unchanged until activate | 1.3 |
| AC-24 | Manual: save scheme with empty name, verify error | 1.7 |
| AC-25 | Manual: save custom provider with empty baseUrl, verify error | 1.7 |
| AC-26 | Manual: save custom provider with no apiType, verify error | 1.7 |
| AC-27 | Manual: save model with empty id, verify error | 1.7 |
| AC-28 | Manual: add model with duplicate id, verify error | 1.7 |
| AC-29 | Manual: add custom provider with duplicate key, verify error | 1.7 |
| AC-30 | Manual: cancel delete dialogs, verify no deletion | 3.1, 3.4 |
| AC-31 | Unit: activate when `%USERPROFILE%/.pi/agent/` doesn't exist | 1.6 |
| AC-32 | Manual: activate to read-only location, verify error message | 1.6 |
| AC-33 | Manual: try adding same built-in twice, verify blocked | 1.7, 3.2 |

## Decisions & Risks

### Decision 1: Storage format
- **Choice**: Single JSON file at `%APPDATA%/pi-mgr/schemes.json` (not SQLite)
- **Rationale**: Small data volume (handful of schemes, tens of providers, hundreds of models max). JSON is debuggable, requires no CGo dependency, and aligns with pi's own JSON config approach. Atomic writes via temp+rename.
- **Risk**: Very large model lists (1000+) could be slow to load. **Mitigation**: Not a v1 concern; model lists are manually entered, practical limits are < 50 per provider.

### Decision 2: Frontend framework
- **Choice**: Vue 3 + TypeScript (Wails default template)
- **Rationale**: Best Wails v2 support, reactive UI for forms, small bundle.

### Decision 3: No import from existing models.json
- Per PRD scope: v1 starts from scratch. Users re-enter config manually.

### Risk: pi models.json format changes
- **Likelihood**: Low (stable public API)
- **Impact**: Medium (serializer would need update)
- **Mitigation**: Serializer is isolated in one file (`serializer.go`); easy to update.

## Unresolved Blockers

None. All 33 ACs are implementable with the described architecture. No product invariants are violated, and all scope boundaries are respected.

Key confirmation: The PRD explicitly excludes `compat`, `headers`, `authHeader`, `oauth`, `modelOverrides`, `cost.tiers`, `thinkingLevelMap`, and all non-Windows platforms. The plan respects all these exclusions.