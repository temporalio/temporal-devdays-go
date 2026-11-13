# One Sandbox Image for every Go Track

TypeScript builds one Sandbox Image per Workshop from one Dockerfile. Go builds
one image, `ghcr.io/temporalio/temporal-devdays-go-sandbox`, with every Workshop
baked in at `/root/workshop/<slug>/`. A Track's `setup-workshop` roots the
editor at its own folder. Java's DevDays repo made the same call (its ADR 0003).

Go's module and build caches are what make the first `go run` and the first
Check take seconds instead of a minute, and nearly all of that cache is the
Temporal SDK and its dependencies, which every Workshop shares. One image warms
it once.

## Considered Options

An image per Workshop, as TypeScript does, isolates a broken Workshop from the
others. Rejected: six images, six digests to re-pin and six pushes per change to
the shared helpers, for content that differs by a few hundred kilobytes of
source.

## Consequences

A change to any Workshop rebuilds the image for every Track, and a Solution
whose tests fail blocks the image for all of them. That is intended: CI proves
every Workshop on every change. Nexus needs extra Namespaces and an Endpoint;
its Track's setup creates them at runtime rather than needing a second image.
