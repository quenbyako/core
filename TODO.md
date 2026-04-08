# TODO

- [ ] Implement readiness checks (based on [Factor XIII](https://github.com/quenbyako/quenbyako/blob/-/guides/cnaa/factor_13.md)) in layout-core runtime module to monitor dependencies like databases and caches.

## Runtime

- [ ] Implement correct construction of resource, without using semconv pacakge. Semconv packages are not compatible between even minor versions, and it should be initialized on application layer.
