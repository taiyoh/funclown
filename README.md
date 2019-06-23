# funclown

This is gorm wrapper for simple usecase.

## Motivation

gorm is powerful and flexible ORM, so each team members use it in his own way. Depending on how it is used, it has security risk.

So, I want to enforce how to use gorm in team development. Hiding gorm operation, enforcing type for arguments.

In funclown, I decided to use Functional Option Pattern for hiding operation.  
Functional Option Pattern has high affinity with `Repository Pattern`. Repository Pattern requires isolated with infrastructure, so need to setting interface for dependency injection. It is the aim for funclown.

For complex query when Read operation, use gorm as it is.
