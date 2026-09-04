# Changelog

## [0.9.0](https://github.com/kropath/kropath-controller/compare/v0.8.0...v0.9.0) (2026-09-04)


### Features

* **KRO-873:** effectiveConfig cascade for EventBridge Pipes fields ([8647822](https://github.com/kropath/kropath-controller/commit/8647822b583335e43342da405e0807f67f0a6f8d))
* **KRO-961:** add SESConfig effectiveConfig cascade ([#89](https://github.com/kropath/kropath-controller/issues/89)) ([9894516](https://github.com/kropath/kropath-controller/commit/98945166590db70e34c7f7517f097f0c21aa8fcf))
* **KRO-973:** add CodeArtifact effectiveConfig cascade support ([dd84566](https://github.com/kropath/kropath-controller/commit/dd845662a7ce9bdf157d9e140c0a0bc456595eaf))
* **KRO-979:** add LoggingTargetPrefix to S3 cascade ([#87](https://github.com/kropath/kropath-controller/issues/87)) ([ffb1900](https://github.com/kropath/kropath-controller/commit/ffb1900131a03e5da2e35f8eda851afbd41f94b7))

## [0.8.0](https://github.com/kropath/kropath-controller/compare/v0.7.0...v0.8.0) (2026-09-02)


### Features

* **KRO-737:** implement CognitoConfig effectiveConfig cascade ([#76](https://github.com/kropath/kropath-controller/issues/76)) ([3038c6f](https://github.com/kropath/kropath-controller/commit/3038c6f61819b2c40d6fb4efabdb112ec302b5e4))
* **KRO-746:** effectiveConfig cascade for Kinesis fields ([#78](https://github.com/kropath/kropath-controller/issues/78)) ([2c26af9](https://github.com/kropath/kropath-controller/commit/2c26af95ca27620c700f0e3935db80f4e24d29a0))
* **KRO-762:** add CloudTrail effectiveConfig cascade ([#79](https://github.com/kropath/kropath-controller/issues/79)) ([5c89a92](https://github.com/kropath/kropath-controller/commit/5c89a92c7ae31b964c8832fa61f1c3e2874e9cea))
* **KRO-772:** add AppScalingConfig effectiveConfig cascade ([#80](https://github.com/kropath/kropath-controller/issues/80)) ([6e3d0cb](https://github.com/kropath/kropath-controller/commit/6e3d0cb2b32bad24e803a8111c080948e0cdfdb8))
* **KRO-790:** add Keyspaces effectiveConfig cascade ([9892c6c](https://github.com/kropath/kropath-controller/commit/9892c6c3a409144c23b575fc76b7ca127109e1a9))
* **KRO-806:** add OpenSearch effectiveConfig cascade controller ([#85](https://github.com/kropath/kropath-controller/issues/85)) ([a72a22e](https://github.com/kropath/kropath-controller/commit/a72a22e423bfc0591266c8337a060fe843d2c928))
* **KRO-814:** add Bedrock effectiveConfig cascade ([#82](https://github.com/kropath/kropath-controller/issues/82)) ([34e5af7](https://github.com/kropath/kropath-controller/commit/34e5af795161ffbf8bec5f112f73c88225cac2a3))
* **KRO-822:** add SageMakerConfig effectiveConfig cascade ([#84](https://github.com/kropath/kropath-controller/issues/84)) ([d5b02aa](https://github.com/kropath/kropath-controller/commit/d5b02aae1a9688ae97c5b697136d9a0d47c91071))
* **KRO-882:** add WAF effectiveConfig cascade ([#83](https://github.com/kropath/kropath-controller/issues/83)) ([b2b3680](https://github.com/kropath/kropath-controller/commit/b2b3680aeba907be3b39512acd63c55a1b1f2624))

## [0.7.0](https://github.com/kropath/kropath-controller/compare/v0.6.0...v0.7.0) (2026-08-28)


### Features

* **KRO-728:** add Route53Config effectiveConfig cascade ([#73](https://github.com/kropath/kropath-controller/issues/73)) ([b7c4e1f](https://github.com/kropath/kropath-controller/commit/b7c4e1f6bbac6d907e5dbb8c947eed41df091d8b))
* **KRO-754:** add SSMConfig effectiveConfig cascade ([#75](https://github.com/kropath/kropath-controller/issues/75)) ([7044d86](https://github.com/kropath/kropath-controller/commit/7044d86644c7c5e5829df7dbfcd1cfea9e019b08))

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
