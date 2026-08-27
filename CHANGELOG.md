# Changelog

## [0.6.0](https://github.com/kropath/kropath-controller/compare/v0.5.0...v0.6.0) (2026-08-27)


### Features

* **KRO-798:** add DSQLConfig effectiveConfig cascade handler ([#72](https://github.com/kropath/kropath-controller/issues/72)) ([57b057e](https://github.com/kropath/kropath-controller/commit/57b057eba91204b5258c74c55ee67ca9e9aac6cf))
* **KRO-857:** effectiveConfig cascade for Athena fields ([ae7ea6c](https://github.com/kropath/kropath-controller/commit/ae7ea6cbb3b07f2f2865e0d2af49e3752c55f974))


### Bug Fixes

* **KRO-894:** pass --metrics-bind-address to chainsaw-start, bind both ports to loopback ([#70](https://github.com/kropath/kropath-controller/issues/70)) ([87b09ed](https://github.com/kropath/kropath-controller/commit/87b09ed6d90e8762b8ba6435d48fef3f359792ef))

## [0.5.0](https://github.com/kropath/kropath-controller/compare/v0.4.0...v0.5.0) (2026-08-24)


### Features

* **KRO-782:** add DocumentDB effectiveConfig cascade ([#64](https://github.com/kropath/kropath-controller/issues/64)) ([d39d78d](https://github.com/kropath/kropath-controller/commit/d39d78dc0924bd01f4b9a66620dd65d89587e3a5))
* **KRO-842:** add GlueConfig effectiveConfig cascade ([#68](https://github.com/kropath/kropath-controller/issues/68)) ([7c2293e](https://github.com/kropath/kropath-controller/commit/7c2293eaff30d43061c54266565162e61cd80716))
* **KRO-849:** add CRD watcher, features active-pending contract, RBAC, and Chainsaw ctrl-dyn-02 and ctrl-dyn-03 ([#66](https://github.com/kropath/kropath-controller/issues/66)) ([70769c9](https://github.com/kropath/kropath-controller/commit/70769c9205f5f11381bb062ca7e1a4d2669ad3e1))
* **KRO-850:** labeloperator convergence + policydocument optional kinds + metrics + Chainsaw ctrl-dyn-04/05 ([#67](https://github.com/kropath/kropath-controller/issues/67)) ([157093e](https://github.com/kropath/kropath-controller/commit/157093ef67583b61734513fadf468dd7f28f44c0))

## [0.4.0](https://github.com/kropath/kropath-controller/compare/v0.3.1...v0.4.0) (2026-08-24)


### Features

* **KRO-716:** add Certificate Manager effectiveConfig cascade ([#59](https://github.com/kropath/kropath-controller/issues/59)) ([76b5552](https://github.com/kropath/kropath-controller/commit/76b5552dfaa87ad350bbeca27f5a0f5efda9b1ea))
* **KRO-830:** implement effectiveConfig cascade handler for EMR fields ([#61](https://github.com/kropath/kropath-controller/issues/61)) ([6d25eb3](https://github.com/kropath/kropath-controller/commit/6d25eb3e8a8efe51e867b9e74e8339937c72805f))
* **KRO-848:** internal/registry + startup discovery gate ([#63](https://github.com/kropath/kropath-controller/issues/63)) ([d018bbb](https://github.com/kropath/kropath-controller/commit/d018bbb6f8d9f09ddfcdfefd738cf2b59fe1637e))
* **KRO-860:** add two-directory CRD fixture support (crds-optional/) ([#62](https://github.com/kropath/kropath-controller/issues/62)) ([bcbade2](https://github.com/kropath/kropath-controller/commit/bcbade22cc9ed6aef09b1bd26d6fa9119e577e33))

## [0.3.1](https://github.com/kropath/kropath-controller/compare/v0.3.0...v0.3.1) (2026-08-19)


### Bug Fixes

* **KRO-675:** match APIGatewayConfig kind casing to the kropath-aws CRD ([#56](https://github.com/kropath/kropath-controller/issues/56)) ([37756b1](https://github.com/kropath/kropath-controller/commit/37756b1def7b83eb94a53cfad17815363d509de9))

## [0.3.0](https://github.com/kropath/kropath-controller/compare/v0.2.0...v0.3.0) (2026-08-17)


### Features

* **KRO-348:** CloudWatch effectiveConfig cascade in kropath-controller ([a2dbc91](https://github.com/kropath/kropath-controller/commit/a2dbc912dfcee01fe34520c60331fd0e6625c4b8))
* **KRO-430:** implement MemoryDB effectiveConfig cascade ([#55](https://github.com/kropath/kropath-controller/issues/55)) ([e241328](https://github.com/kropath/kropath-controller/commit/e24132800d01a770fc399647f237468632d01bfa))
* **KRO-571:** add MSK effectiveConfig cascade ([#54](https://github.com/kropath/kropath-controller/issues/54)) ([ae8e9a1](https://github.com/kropath/kropath-controller/commit/ae8e9a1111bc582c586e652d8cbb2ce1b6898f18))

## [0.2.0](https://github.com/kropath/kropath-controller/compare/v0.1.0...v0.2.0) (2026-08-17)


### Features

* **KRO-555:** add APIGatewayConfig effectiveConfig cascade ([#49](https://github.com/kropath/kropath-controller/issues/49)) ([3b33d16](https://github.com/kropath/kropath-controller/commit/3b33d166a3075290d07859f86e53803fee4ae027))

## [0.1.0](https://github.com/kropath/kropath-controller/compare/v0.0.1...v0.1.0) (2026-08-14)


### Features

* **KRO-641:** add pr-title lint, OCI labels, and release process docs ([#46](https://github.com/kropath/kropath-controller/issues/46)) ([02fd4a0](https://github.com/kropath/kropath-controller/commit/02fd4a054ac8cd2ac67cc92cd7a16deb462a85e8))
* **KRO-656:** implement controller gaps 1-3 — dynamic global namespace, local default KPC, S3 parity ([#48](https://github.com/kropath/kropath-controller/issues/48)) ([1bbb71f](https://github.com/kropath/kropath-controller/commit/1bbb71fd4b49a09209f5da4f690968334689656c))
