# Repository Interface Extraction Blueprint — T1.1

This document catalogs the repository interfaces found in `internal/repository/` and defines the conventions for extracting them during Phase 1 (T1.2–T1.5). It is read-only analysis output; no code changes are included here.

## Section 1 — Interface Inventory

### 1. Admin / User Repository

- **Source file:** `admin_repository.go`
- **Repository interface:** `UserRepository`
- **Interface name:** `UserRepository`
- **Concrete type name:** `MongoUserRepository`
- **Constructor signature:** `func NewMongoUserRepository(db *mongo.Database) *MongoUserRepository`
- **Method count:** 9
- **Method signatures:**
  - `FindByEmailHash(ctx context.Context, emailHash string) (*models.User, error)`
  - `EmailExistsByHash(ctx context.Context, emailHash string) (bool, error)`
  - `Create(ctx context.Context, user models.User) error`
  - `FindByID(ctx context.Context, id string) (*models.User, error)`
  - `UpdateProfile(ctx context.Context, id string, username string, passwordHash *string) error`
  - `BeginMFAEnrollment(ctx context.Context, id string, secretEnc string, recoveryHashes []string) error`
  - `EnableMFA(ctx context.Context, id string) error`
  - `DisableMFA(ctx context.Context, id string) error`
  - `ReplaceRecoveryCodes(ctx context.Context, id string, hashes []string) error`
- **Collections used:** `users`

### 2. Incident Repository

- **Source file:** `incident_repository.go`
- **Repository interface:** `IncidentRepository`
- **Interface name:** `IncidentRepository`
- **Concrete type name:** `MongoIncidentRepository`
- **Constructor signature:** `func NewMongoIncidentRepository(db *mongo.Database) *MongoIncidentRepository`
- **Method count:** 12
- **Method signatures:**
  - `List(ctx context.Context, filter bson.M, page, limit int) ([]models.Incident, int64, error)`
  - `FindByID(ctx context.Context, id primitive.ObjectID) (models.Incident, error)`
  - `InsertIncident(ctx context.Context, incident models.Incident) error`
  - `UpdateIncidentByID(ctx context.Context, id primitive.ObjectID, setFields bson.M) (models.Incident, error)`
  - `DeleteIncidentByID(ctx context.Context, id primitive.ObjectID) error`
  - `InsertUpdate(ctx context.Context, update models.IncidentUpdate) error`
  - `ApplyIncidentStatus(ctx context.Context, incidentID primitive.ObjectID, status models.IncidentStatus) error`
  - `ListUpdates(ctx context.Context, incidentID primitive.ObjectID) ([]models.IncidentUpdate, error)`
  - `InsertAuditLog(ctx context.Context, audit models.AuditLog) error`
  - `ListHistory(ctx context.Context, incidentID primitive.ObjectID) ([]models.AuditLog, error)`
  - `CountComponents(ctx context.Context, ids []primitive.ObjectID) (int64, error)`
  - `CountSubComponentsByComponent(ctx context.Context, componentID primitive.ObjectID, ids []primitive.ObjectID) (int64, error)`
- **Collections used:** `incidents`, `incident_updates`, `audit_logs`, `components`, `subcomponents`

### 3. Maintenance Repository

- **Source file:** `maintenance_repository.go`
- **Repository interface:** `MaintenanceRepository`
- **Interface name:** `MaintenanceRepository`
- **Concrete type name:** `MongoMaintenanceRepository`
- **Constructor signature:** `func NewMongoMaintenanceRepository(db *mongo.Database) *MongoMaintenanceRepository`
- **Method count:** 8
- **Method signatures:**
  - `List(ctx context.Context, page, limit int) ([]models.Maintenance, int64, error)`
  - `ListPublic(ctx context.Context, page, limit int) ([]models.Maintenance, int64, error)`
  - `FindByID(ctx context.Context, id primitive.ObjectID) (models.Maintenance, error)`
  - `Insert(ctx context.Context, maintenance models.Maintenance) error`
  - `UpdateByID(ctx context.Context, id primitive.ObjectID, setFields bson.M) (models.Maintenance, error)`
  - `DeleteByID(ctx context.Context, id primitive.ObjectID) error`
  - `InsertAuditLog(ctx context.Context, audit models.AuditLog) error`
  - `ListHistory(ctx context.Context, maintenanceID primitive.ObjectID) ([]models.AuditLog, error)`
- **Collections used:** `maintenance`, `audit_logs`

### 4. Monitor Repository

- **Source file:** `monitor_repository.go`
- **Repository interface:** `MonitorRepository`
- **Interface name:** `MonitorRepository`
- **Concrete type name:** `MongoMonitorRepository`
- **Constructor signature:** `func NewMongoMonitorRepository(db *mongo.Database) *MongoMonitorRepository`
- **Method count:** 9
- **Method signatures:**
  - `Insert(ctx context.Context, monitor models.Monitor) error`
  - `Update(ctx context.Context, id primitive.ObjectID, monitor models.Monitor) (bool, error)`
  - `Delete(ctx context.Context, id primitive.ObjectID) (bool, error)`
  - `List(ctx context.Context, page, limit int) ([]models.Monitor, int64, error)`
  - `ListLogs(ctx context.Context, monitorID primitive.ObjectID, limit int64) ([]models.MonitorLog, error)`
  - `FindLogsByMonitorIDPaginated(ctx context.Context, monitorID primitive.ObjectID, page, limit int) ([]models.MonitorLog, int64, error)`
  - `ListUptime(ctx context.Context, monitorID primitive.ObjectID, since time.Time) ([]models.DailyUptime, error)`
  - `ListOutages(ctx context.Context) ([]models.Outage, error)`
  - `ListHistory(ctx context.Context, monitorID primitive.ObjectID, limit int64) ([]models.EnhancedMonitorLog, error)`
- **Collections used:** `monitors`, `monitor_logs`, `daily_uptime`, `outages`, `enhanced_monitor_logs`

### 5. Settings Repository

- **Source file:** `settings_repository.go`
- **Repository interface:** `SettingsRepository`
- **Interface name:** `SettingsRepository`
- **Concrete type name:** `MongoSettingsRepository`
- **Constructor signature:** `func NewMongoSettingsRepository(db *mongo.Database) *MongoSettingsRepository`
- **Method count:** 2
- **Method signatures:**
  - `GetSSOSettings(ctx context.Context) (*models.StatusPageSSOSettings, error)`
  - `UpdateSSOSettings(ctx context.Context, updates bson.M) (*models.StatusPageSSOSettings, error)`
- **Collections used:** `settings`

### 6. Status Repository

- **Source file:** `status_repository.go`
- **Repository interface:** `StatusRepository`
- **Interface name:** `StatusRepository`
- **Concrete type name:** `MongoStatusRepository`
- **Constructor signature:** `func NewMongoStatusRepository(db *mongo.Database) *MongoStatusRepository`
- **Method count:** 20
- **Method signatures:**
  - `ListComponents(ctx context.Context) ([]models.Component, error)`
  - `ListComponentsByIDs(ctx context.Context, ids []primitive.ObjectID) ([]models.Component, error)`
  - `ListSubComponentsByComponentIDs(ctx context.Context, componentIDs []primitive.ObjectID) ([]models.SubComponent, error)`
  - `ListAllSubComponents(ctx context.Context) ([]models.SubComponent, error)`
  - `ListSubComponentsByIDs(ctx context.Context, ids []primitive.ObjectID) ([]models.SubComponent, error)`
  - `FindMonitorByID(ctx context.Context, id primitive.ObjectID) (*models.Monitor, error)`
  - `FindMonitorBySubComponentID(ctx context.Context, subComponentID primitive.ObjectID) (*models.Monitor, error)`
  - `ListMonitorsByTargets(ctx context.Context, componentIDs []primitive.ObjectID, subComponentIDs []primitive.ObjectID) ([]models.Monitor, error)`
  - `ListMonitorsByServiceID(ctx context.Context, serviceID primitive.ObjectID) ([]models.Monitor, error)`
  - `ListMonitorLogsByMonitorIDsSince(ctx context.Context, monitorIDs []primitive.ObjectID, since time.Time) ([]models.MonitorLog, error)`
  - `ListDailyUptimeSinceByMonitorIDs(ctx context.Context, monitorIDs []primitive.ObjectID, since time.Time) ([]models.DailyUptime, error)`
  - `ListActiveIncidents(ctx context.Context) ([]models.Incident, error)`
  - `ListActiveMaintenanceAt(ctx context.Context, at time.Time) ([]models.Maintenance, error)`
  - `CountActiveIncidents(ctx context.Context) (int64, error)`
  - `CountActiveMaintenanceAt(ctx context.Context, at time.Time) (int64, error)`
  - `FindLatestIncidentByComponent(ctx context.Context, componentID primitive.ObjectID) (*models.Incident, error)`
  - `ListIncidentsByCreatedAtRange(ctx context.Context, start, end time.Time) ([]models.Incident, error)`
  - `ListResolvedIncidentsSince(ctx context.Context, since time.Time) ([]models.Incident, error)`
  - `ListIncidentsByAffectedComponents(ctx context.Context, affectedIDs []primitive.ObjectID, limit int64) ([]models.Incident, error)`
  - `ListIncidentUpdatesByIncidentIDs(ctx context.Context, incidentIDs []primitive.ObjectID) (map[primitive.ObjectID][]models.IncidentUpdate, error)`
- **Collections used:** `components`, `subcomponents`, `monitors`, `monitor_logs`, `daily_uptime`, `incidents`, `incident_updates`, `maintenance`

### 7. Subcomponent Repository

- **Source file:** `subcomponent_repository.go`
- **Repository interface:** `SubComponentRepository`
- **Interface name:** `SubComponentRepository`
- **Concrete type name:** `MongoSubComponentRepository`
- **Constructor signature:** `func NewMongoSubComponentRepository(db *mongo.Database) *MongoSubComponentRepository`
- **Method count:** 8
- **Method signatures:**
  - `List(ctx context.Context, filter bson.M, page, limit int) ([]models.SubComponent, int64, error)`
  - `Insert(ctx context.Context, sub models.SubComponent) error`
  - `UpdateByID(ctx context.Context, id primitive.ObjectID, setFields bson.M) (models.SubComponent, error)`
  - `FindByID(ctx context.Context, id primitive.ObjectID) (models.SubComponent, error)`
  - `DeleteByID(ctx context.Context, id primitive.ObjectID) (int64, error)`
  - `CountByComponentID(ctx context.Context, componentID primitive.ObjectID) (int64, error)`
  - `ComponentExists(ctx context.Context, id primitive.ObjectID) (bool, error)`
  - `CleanupReferencesForDeletedSubComponent(ctx context.Context, subComponentID primitive.ObjectID, componentID primitive.ObjectID) error`
- **Collections used:** `subcomponents`, `components`, `monitors`, `outages`, `incidents`

### 8. Subscriber Repository

- **Source file:** `subscriber_repository.go`
- **Repository interface:** `SubscriberRepository`
- **Interface name:** `SubscriberRepository`
- **Concrete type name:** `MongoSubscriberRepository`
- **Constructor signature:** `func NewMongoSubscriberRepository(db *mongo.Database) *MongoSubscriberRepository`
- **Method count:** 4
- **Method signatures:**
  - `FindByEmail(ctx context.Context, email string) (*models.Subscriber, error)`
  - `Insert(ctx context.Context, sub models.Subscriber) error`
  - `List(ctx context.Context, page, limit int) ([]models.Subscriber, int64, error)`
  - `DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error)`
- **Collections used:** `subscribers`

### 9. Webhook Repository

- **Source file:** `webhook_repository.go`
- **Repository interface:** `WebhookChannelRepository`
- **Interface name:** `WebhookChannelRepository`
- **Concrete type name:** `MongoWebhookChannelRepository`
- **Constructor signature:** `func NewMongoWebhookChannelRepository(db *mongo.Database) *MongoWebhookChannelRepository`
- **Method count:** 3
- **Method signatures:**
  - `List(ctx context.Context, page, limit int) ([]models.WebhookChannel, int64, error)`
  - `Insert(ctx context.Context, channel models.WebhookChannel) error`
  - `DeleteByID(ctx context.Context, id primitive.ObjectID) (bool, error)`
- **Collections used:** `webhook_channels`

## Section 2 — Naming Inconsistencies Found

| # | File | Current name | Issue | Recommended resolution | Rationale |
|---|---|---|---|---|---|
| 1 | `admin_repository.go` | `UserRepository` | File name says Admin, interface says User | Keep `UserRepository`; rename file to `user_repository.go` in a later cleanup slice. | The interface matches the `User` model and the auth service domain. The file name is the mismatch. |
| 2 | `subcomponent_repository.go` | `SubComponentRepository` | Uses `SubComponent` casing, while models and package names use `Subcomponent` | Rename interface to `SubcomponentRepository`. | The Go package is `subcomponent`, the model is `SubComponent`, and the domain term is `Subcomponent`. Interface names should follow package-level domain spelling. |
| 3 | `webhook_repository.go` | `WebhookChannelRepository` | File name says Webhook, interface says WebhookChannel | Keep `WebhookChannelRepository`; rename file to `webhook_channel_repository.go` in a later cleanup slice. | The model is `WebhookChannel` and the concrete type is `MongoWebhookChannelRepository`. The file name is the mismatch. |

## Section 3 — Interface Convention Specification

These rules apply to T1.2–T1.5 when extracting the Go interfaces.

1. **Interface naming**
   - Use `<Domain>Repository` where `<Domain>` matches the service domain and the primary model name.
   - Resolve the Section 2 inconsistencies first: `UserRepository`, `SubcomponentRepository`, `WebhookChannelRepository`.

2. **File placement**
   - Keep interfaces in the existing flat package `internal/repository`.
   - Do not create `internal/repository/interfaces/` unless a later phase explicitly requires it. The current codebase already places interfaces and concrete implementations in the same file, so a single flat package is the smallest change and matches the existing structure.

3. **Constructor return-type rule**
   - Every constructor must return the interface type, not the concrete pointer:
     ```go
     func New<Domain>Repository(db *mongo.Database) <Domain>Repository
     ```
   - The concrete type (`Mongo<Domain>Repository`) remains unexported except for the constructor. This is the change that enables service-layer dependency inversion.

4. **Compile-time assertion rule**
   - Every concrete type must include a compile-time interface satisfaction check:
     ```go
     var _ <Domain>Repository = (*Mongo<Domain>Repository)(nil)
     ```

5. **Parameter conventions**
   - Every method accepts `ctx context.Context` as the first parameter.
   - Every method returns a result value and an error (`(result, error)`). Error-only returns are not used in this repository; keep the existing pattern.
   - Entity identifiers use `primitive.ObjectID` for MongoDB-backed repositories. `string` ids are only acceptable where the existing codebase already uses them (for example, `UserRepository.FindByID` and `UserRepository.UpdateProfile`).
   - Paginated list methods use `(page, limit int)`.

6. **Cross-collection access notes**
   - The following methods read or write collections outside their primary domain. T1.2–T1.5 must preserve these couplings; they are not errors, but they are documented coupling points for future refactoring.
     - `IncidentRepository.CountComponents` reads `components`.
     - `IncidentRepository.CountSubComponentsByComponent` reads `subcomponents`.
     - `IncidentRepository.DeleteIncidentByID` deletes from `incident_updates`.
     - `MaintenanceRepository.InsertAuditLog` and `ListHistory` read/write `audit_logs`.
     - `MonitorRepository.ListLogs` reads `monitor_logs`.
     - `MonitorRepository.FindLogsByMonitorIDPaginated` reads `monitor_logs`.
     - `MonitorRepository.ListUptime` reads `daily_uptime`.
     - `MonitorRepository.ListOutages` reads `outages`.
     - `MonitorRepository.ListHistory` reads `enhanced_monitor_logs`.
     - `StatusRepository` is intentionally cross-collection: it reads `components`, `subcomponents`, `monitors`, `monitor_logs`, `daily_uptime`, `incidents`, `incident_updates`, and `maintenance` to build public status views.
     - `SubcomponentRepository.ComponentExists` reads `components`.
     - `SubcomponentRepository.CleanupReferencesForDeletedSubComponent` updates `monitors`, `outages`, and `incidents`.
   - These cross-collection touches are a consequence of the current MongoDB-first design and the unified `StatusRepository` aggregation role. Addressing them is out of scope for T1.1.

7. **Future-direction note on driver-agnostic interfaces**
   - The current repository interfaces leak driver-specific types in two ways: `bson.M` filters and `primitive.ObjectID` identifiers. T1.2–T1.5 must catalog these signatures as they exist today. A later phase may introduce driver-agnostic parameter types if SQL driver support becomes active, but that conversion is not part of T1.1.

8. **Service-layer dependency rule**
   - Services must depend on repository interfaces only (FR-049, ADR-2, ADR-11). After T1.2–T1.5, service constructors must be updated from `NewService(repo *Mongo<Domain>Repository)` to `NewService(repo <Domain>Repository)`. No handler may instantiate a repository or call `database.GetDB()` directly.

## Section 4 — Constructor Injection Template

```go
package repository

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"

    "github.com/fresp/Statora/internal/models"
)

// <Domain>Repository defines the persistence contract for <domain> operations.
type <Domain>Repository interface {
    // MethodName(ctx context.Context, ...) (...result, error)
}

// Mongo<Domain>Repository is the MongoDB-backed implementation of <Domain>Repository.
type Mongo<Domain>Repository struct {
    collection *mongo.Collection
}

// New<Domain>Repository returns a MongoDB-backed <Domain>Repository.
// It accepts *mongo.Database directly. Service constructors will receive
// the interface, wired by the RepositoryFactory in T1.5.
func New<Domain>Repository(db *mongo.Database) <Domain>Repository {
    return &Mongo<Domain>Repository{
        collection: db.Collection("<collection_name>"),
    }
}

// Compile-time interface satisfaction check.
var _ <Domain>Repository = (*Mongo<Domain>Repository)(nil)
```

Replace `<Domain>` and `<collection_name>` with the appropriate domain and primary collection name. If a repository needs multiple collections, add them as fields to `Mongo<Domain>Repository` and initialize them in the constructor. Keep constructor injection pure: no global state, no DI framework, no side effects beyond the constructor body.
