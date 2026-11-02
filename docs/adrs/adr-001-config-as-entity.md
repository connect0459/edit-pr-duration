# ADR-001: Treat Config as an Entity

## Status

- [ ] Proposed
- [x] Accepted
- [ ] Deprecated

## Context

While designing edit-pr-duration, we had to decide which category of the domain layer the application configuration (Config) belongs to.

### Factors Considered

1. **Whether it is persisted**
   - Config is persisted as a JSON file (config.json)
   - The file path acts as an implicit identifier

2. **From the DDD (Domain-Driven Design) perspective**
   - The traditional definition says an Entity "has an identity and is mutable", while a Value Object "has no identity and is immutable"
   - However, under the **Always-Valid Domain Model**, an Entity should also be immutable

3. **Go idioms**
   - Structs in Go are value types, and keeping them immutable is recommended
   - Changes are expressed by creating a new instance

4. **Presence of domain logic**
   - Config contains business logic such as IsWorkday(), WorkStartTime(), and WorkEndTime()
   - This is pure domain knowledge

### Options

1. **entities/config.go** - Treat it as an Entity
   - Rationale: the file path is the identifier, it is persisted, and it contains domain logic
   - Follows the Always-Valid Domain Model principle: an immutable Entity

2. **valueobjects/config.go** - Treat it as a Value Object (the existing placement)
   - Rationale: no explicit ID field; equality is judged by attribute values
   - Problem: it cannot explain why the object is persisted

3. **configuration/config.go** - A dedicated category
   - Rationale: Config is special enough to go beyond the Entity/Value Object dichotomy
   - Problem: the complexity of introducing a new category

## Decision

**Treat Config as an Entity and place it in `internal/domain/entities/config.go`.**

### Implementation Policy

1. **Separating Config**
   - The Config itself → `entities/config.go` (Entity)
   - Period, WorkHours, Options → `valueobjects/*.go` (Value Objects)

2. **Implementing immutability**
   - All fields are unexported (lowercase)
   - Changing a value is done by creating a new instance (see `main.go`)

   ```go
   config = entities.NewConfig(
       config.Repositories(),
       config.Period(),
       config.WorkHours(),
       config.Holidays(),
       config.ReplacementPatterns(),
       config.Author(),
       valueobjects.Options{DryRun: true, ...},
   )
   ```

3. **Identifier**
   - The file path (of the JSON file) is the implicit identifier
   - It is identified through ConfigRepository's Load(path string)

### Rationale

- **Always-Valid Domain Model**: Entities should also be immutable
- **Rich Domain Object**: Data + business logic (avoids the Anemic Domain Model)
- **DDD**: It is natural to treat persisted objects as Entities

## Consequences

### Positive Impact

1. **Clearer architecture**
   - The distinction between Entities (Config, PRInfo) and Value Objects (Period, WorkHours, Options) is clear
   - The responsibilities of the domain layer are organized

2. **Better tests**
   - The tests in application/service_test.go pass (a regex pattern mismatch was also fixed)
   - Objects are created explicitly via entities.NewConfig()

3. **Improved maintainability**
   - Domain logic (IsWorkday, etc.) is consolidated in Config
   - Value Objects (Period, etc.) stay independent as pure values

### Negative Impact

1. **Refactoring cost**
   - All import paths had to be updated (valueobjects → entities)
   - Eight files needed changes

2. **Learning cost**
   - The idea that "an Entity is also immutable" may differ from the traditional understanding of DDD
   - Understanding the Always-Valid Domain Model is required

## References

- [Always-Valid Domain Model](https://enterprisecraftsmanship.com/posts/always-valid-domain-model/)
- [DDD: Entity vs Value Object](https://martinfowler.com/bliki/EvansClassification.html)
- [Anemic Domain Model (anti-pattern)](https://martinfowler.com/bliki/AnemicDomainModel.html)

## Related File Paths

### Initial implementation (2025-12-19)

- internal/domain/entities/config.go (new)
- internal/domain/entities/prinfo.go (existing)
- internal/domain/valueobjects/period.go (new)
- internal/domain/valueobjects/workhours.go (new)
- internal/domain/valueobjects/options.go (new)
- internal/domain/valueobjects/config.go (deleted)

### Import path updates (2025-12-19)

- internal/domain/repositories/config_repository.go (valueobjects → entities)
- internal/infrastructure/json/config_repository.go (valueobjects → entities)
- internal/application/service.go (valueobjects → entities)
- internal/application/service_test.go (valueobjects → entities, test data fixed)
- internal/domain/services/calculator.go (valueobjects → entities)
- main.go (valueobjects → entities)
