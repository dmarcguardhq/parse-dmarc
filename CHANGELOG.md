# Changelog

## [1.7.0](https://github.com/dmarcguardhq/parse-dmarc/compare/v1.6.1...v1.7.0) (2026-09-18)


### Features

* log attachments rejected as non-DMARC ([#208](https://github.com/dmarcguardhq/parse-dmarc/issues/208)) ([31b6a23](https://github.com/dmarcguardhq/parse-dmarc/commit/31b6a230fe56c749296dcac7b487675651b852a6))


### Bug Fixes

* accept DMARC reports sent as inline MIME parts ([#207](https://github.com/dmarcguardhq/parse-dmarc/issues/207)) ([20b2be0](https://github.com/dmarcguardhq/parse-dmarc/commit/20b2be03d8a43197eabde63eff452e276eaf2b63))
* **docs:** update brew installation ref ([d62606e](https://github.com/dmarcguardhq/parse-dmarc/commit/d62606e99b80c5eddfe1b2d33c5aaed2631ee0aa))


### Chores

* **docs:** publish brew to the new org-wide repository ([432f010](https://github.com/dmarcguardhq/parse-dmarc/commit/432f010856838969b99a078339d72830421a9678))

## [1.6.1](https://github.com/dmarcguardhq/parse-dmarc/compare/v1.6.0...v1.6.1) (2026-09-15)


### Bug Fixes

* **CI:** publish the image to the new repo on main ([f89a9ec](https://github.com/dmarcguardhq/parse-dmarc/commit/f89a9ece752e3de0bf5c42f3583f41209cf46d55))
* **docker:** publish to the new repository ([c29118f](https://github.com/dmarcguardhq/parse-dmarc/commit/c29118feff7ce3b2cb4c045b6dabb2e621b7c050)), closes [#201](https://github.com/dmarcguardhq/parse-dmarc/issues/201)
* **docs:** remove report card ([6fc3871](https://github.com/dmarcguardhq/parse-dmarc/commit/6fc3871905dfb970558dfe28e30830aa72548c13))
* **docs:** update docker-run snippet to the new image repo ([38d8e59](https://github.com/dmarcguardhq/parse-dmarc/commit/38d8e59914b3cdef25165b667e8414cd5a71fa5d))


### Chores

* **deps:** lock file maintenance ([#198](https://github.com/dmarcguardhq/parse-dmarc/issues/198)) ([a4d5d96](https://github.com/dmarcguardhq/parse-dmarc/commit/a4d5d96a7b4e69392babefbe0541a25e4001a640))
* **docs:** rewrite README around Parse DMARC, split metrics and MCP docs, credit DMARCguard ([#196](https://github.com/dmarcguardhq/parse-dmarc/issues/196)) ([aed41c1](https://github.com/dmarcguardhq/parse-dmarc/commit/aed41c1f1a7bd37e13d1276bbc8b3db5d0a2299d))

## [1.6.0](https://github.com/dmarcguardhq/dmarcguard/compare/v1.5.7...v1.6.0) (2026-09-11)


### Features

* **imap:** support internal CA, skip-verify and STARTTLS ([#194](https://github.com/dmarcguardhq/dmarcguard/issues/194)) ([98abb52](https://github.com/dmarcguardhq/dmarcguard/commit/98abb526d856a2c560fabc090a383a206fcec2cc)), closes [#193](https://github.com/dmarcguardhq/dmarcguard/issues/193)

## [1.5.7](https://github.com/dmarcguardhq/dmarcguard/compare/v1.5.6...v1.5.7) (2026-09-05)


### Bug Fixes

* **dashboard:** judge health over delivered mail, in the backend ([#191](https://github.com/dmarcguardhq/dmarcguard/issues/191)) ([53d8944](https://github.com/dmarcguardhq/dmarcguard/commit/53d8944fdc95ed4199e3ad47505e461e74851d5a))
* **dashboard:** show null reports as EMPTY instead of FAIL ([#187](https://github.com/dmarcguardhq/dmarcguard/issues/187)) ([dcf9d98](https://github.com/dmarcguardhq/dmarcguard/commit/dcf9d98cff2ebda51c6157bcb603f0d9550941e2))
* **ingest:** bound report decompression to stop a gzip/zip bomb ([#189](https://github.com/dmarcguardhq/dmarcguard/issues/189)) ([f50ee41](https://github.com/dmarcguardhq/dmarcguard/commit/f50ee4129f8f56fd8a1647f9859978f73e778d41))
* **precommit:** run prettier on changed files only ([729f64d](https://github.com/dmarcguardhq/dmarcguard/commit/729f64db6044da83bd7bfc356b926a8e05d062cf))


### Build & Dependencies

* **deps:** update actions/setup-go action to v7 ([#175](https://github.com/dmarcguardhq/dmarcguard/issues/175)) ([99e4a48](https://github.com/dmarcguardhq/dmarcguard/commit/99e4a48f9e04e854c09b2ecd906ed543131c6432))
* **deps:** update golang docker tag to v1.27 ([#184](https://github.com/dmarcguardhq/dmarcguard/issues/184)) ([afdb202](https://github.com/dmarcguardhq/dmarcguard/commit/afdb202e842fcddaf88d6f54dcbeb779dd4ac9d8))
* **deps:** update module github.com/coreos/go-oidc/v3 to v3.21.0 ([#188](https://github.com/dmarcguardhq/dmarcguard/issues/188)) ([675492d](https://github.com/dmarcguardhq/dmarcguard/commit/675492dfb372adcb03b5b5b56db9c367e85ddcaa))
* **deps:** update module github.com/modelcontextprotocol/go-sdk to v1.7.0 ([#178](https://github.com/dmarcguardhq/dmarcguard/issues/178)) ([f10b023](https://github.com/dmarcguardhq/dmarcguard/commit/f10b0233751ebdaa1e9d94a28906fb69ac417672))
* **deps:** update module github.com/urfave/cli/v3 to v3.11.0 ([#181](https://github.com/dmarcguardhq/dmarcguard/issues/181)) ([e5d3fef](https://github.com/dmarcguardhq/dmarcguard/commit/e5d3fefa373a85687bae2e13421e04308e0a7f44))
* **deps:** update module modernc.org/sqlite to v1.58.0 ([#179](https://github.com/dmarcguardhq/dmarcguard/issues/179)) ([21daa5b](https://github.com/dmarcguardhq/dmarcguard/commit/21daa5b8c797968a5d3c9d51533cc41138082843))
* **deps:** update non-major dependencies ([#174](https://github.com/dmarcguardhq/dmarcguard/issues/174)) ([4bae98f](https://github.com/dmarcguardhq/dmarcguard/commit/4bae98f85398c2d160401c64c4f37a0bdab5f349))
* **deps:** upgrade JS dependencies (ky v2, pinia v4, sharp 0.35, vite 8.2.2) ([#192](https://github.com/dmarcguardhq/dmarcguard/issues/192)) ([d80f128](https://github.com/dmarcguardhq/dmarcguard/commit/d80f1289638e230e2193ab8e1f3b349341679411))


### Chores

* **CI:** enrich release changelog ([5f9f9e8](https://github.com/dmarcguardhq/dmarcguard/commit/5f9f9e85e69a0816176df62f2619e6291168d12b))
* **deps:** lock file maintenance ([#177](https://github.com/dmarcguardhq/dmarcguard/issues/177)) ([b3da7f2](https://github.com/dmarcguardhq/dmarcguard/commit/b3da7f2d067e413323580ef20f6eb4b17b99a64e))

## [1.5.6](https://github.com/dmarcguardhq/dmarcguard/compare/v1.5.5...v1.5.6) (2026-08-18)


### Bug Fixes

* **imap:** unwrap nested message/rfc822 report attachments ([c82f118](https://github.com/dmarcguardhq/dmarcguard/commit/c82f118c375481956cccd352ec00c3275aa01e0a))

## [1.5.5](https://github.com/dmarcguardhq/dmarcguard/compare/v1.5.4...v1.5.5) (2026-07-22)


### Bug Fixes

* **docker:** make /data writable under non-root runtimes ([#171](https://github.com/dmarcguardhq/dmarcguard/issues/171)) ([90a36d3](https://github.com/dmarcguardhq/dmarcguard/commit/90a36d3359edf9dd7e2529b412261c3ecbfe67f7))

## [1.5.4](https://github.com/dmarcguardhq/dmarcguard/compare/v1.5.3...v1.5.4) (2026-06-20)


### Bug Fixes

* accept fetch-interval without prefix ([06fd520](https://github.com/dmarcguardhq/dmarcguard/commit/06fd520d1a816795d8e35f6a7c9b1b5eed5c096c)), closes [#162](https://github.com/dmarcguardhq/dmarcguard/issues/162)
* **CI:** name docker repo explicitly ([0ebbf87](https://github.com/dmarcguardhq/dmarcguard/commit/0ebbf87a6aad51ad735434982734105d79a97893))
* **CI:** publish to ghcr.io as well ([9c0bf28](https://github.com/dmarcguardhq/dmarcguard/commit/9c0bf284a08c50487bb6e8fe475b0248a836cedc))
* **CI:** publish to new and old docker hub repo ([e7eda52](https://github.com/dmarcguardhq/dmarcguard/commit/e7eda52d98e7d3d7d127f199f0f06376007e5cb4))
* **CI:** remove docker repo from old ghcr.io ([d3f24c0](https://github.com/dmarcguardhq/dmarcguard/commit/d3f24c00961876ac7bd8a976a16ada168a59be28))
* **deps:** update module github.com/caarlos0/env/v11 to v11.4.1 ([#155](https://github.com/dmarcguardhq/dmarcguard/issues/155)) ([31398d3](https://github.com/dmarcguardhq/dmarcguard/commit/31398d31413b2c204fb4fd0946bc3d087a490b77))
* **deps:** update module github.com/mattn/go-sqlite3 to v1.14.38 ([#129](https://github.com/dmarcguardhq/dmarcguard/issues/129)) ([40fb4d4](https://github.com/dmarcguardhq/dmarcguard/commit/40fb4d4a9f2048af86f3cf77be12df5a16cf134b))
* **deps:** update module github.com/modelcontextprotocol/go-sdk to v1.6.1 ([#151](https://github.com/dmarcguardhq/dmarcguard/issues/151)) ([eb1fb95](https://github.com/dmarcguardhq/dmarcguard/commit/eb1fb953b6e91199a7de4df959802a2dffdeb9ef))
* move cli init to cmd/ and grab values using Destination syntax ([e04af1a](https://github.com/dmarcguardhq/dmarcguard/commit/e04af1aefd1c33c74780afd2b4e00805a951c1b0))

## [1.5.3](https://github.com/meysam81/parse-dmarc/compare/v1.5.2...v1.5.3) (2026-03-28)


### Bug Fixes

* **deps:** update module github.com/urfave/cli/v3 to v3.8.0 ([#135](https://github.com/meysam81/parse-dmarc/issues/135)) ([dc06765](https://github.com/meysam81/parse-dmarc/commit/dc06765f9d65c0dd67a2debcb12ed3499813ed05))
* **docker:** create DB parent dirs ([e8d6497](https://github.com/meysam81/parse-dmarc/commit/e8d6497227119e4e7e0d348f10386e4b657b630d)), closes [#138](https://github.com/meysam81/parse-dmarc/issues/138)

## [1.5.2](https://github.com/meysam81/parse-dmarc/compare/v1.5.1...v1.5.2) (2026-03-13)


### Bug Fixes

* **deps:** update module github.com/modelcontextprotocol/go-sdk to v1.4.1 ([#127](https://github.com/meysam81/parse-dmarc/issues/127)) ([ce63d91](https://github.com/meysam81/parse-dmarc/commit/ce63d9146a8865e1f1bf5869b6502b6f62c30dc5))
* use PEEK when mark-seen is set to false ([#125](https://github.com/meysam81/parse-dmarc/issues/125)) ([e54e269](https://github.com/meysam81/parse-dmarc/commit/e54e26900faed10f916afed47d23b3a2a1f65467))

## [1.5.1](https://github.com/meysam81/parse-dmarc/compare/v1.5.0...v1.5.1) (2026-03-13)


### Bug Fixes

* **CI:** remove mcp entry from goreleaser altogether ([1d769a4](https://github.com/meysam81/parse-dmarc/commit/1d769a45d05ea5afc5340e87fb60cec3fe3514ce))

## [1.5.0](https://github.com/meysam81/parse-dmarc/compare/v1.4.8...v1.5.0) (2026-03-13)


### Features

* allow customizing SEEN & move-folder behavior after processing ([80346df](https://github.com/meysam81/parse-dmarc/commit/80346dfddc56844dd59c9474fb788ca1d0d9312c)), closes [#118](https://github.com/meysam81/parse-dmarc/issues/118)


### Bug Fixes

* **CI:** disable publishing to mcp registry ([c2e3969](https://github.com/meysam81/parse-dmarc/commit/c2e3969e90554479921437be4df0d011b4214a89))
* **CI:** use oxc the default vite minifier ([2cc92e1](https://github.com/meysam81/parse-dmarc/commit/2cc92e1d85734bc1cc34b91b209eb66b13336b5d))
* **deps:** update module github.com/caarlos0/env/v11 to v11.4.0 ([#104](https://github.com/meysam81/parse-dmarc/issues/104)) ([e14e6d2](https://github.com/meysam81/parse-dmarc/commit/e14e6d2d7e8ecbfd8d7e5564c600b165b4135946))
* **deps:** update module github.com/goccy/go-json to v0.10.6 ([#119](https://github.com/meysam81/parse-dmarc/issues/119)) ([7ad51d3](https://github.com/meysam81/parse-dmarc/commit/7ad51d3a0b0395d99b542293ae5dd2a0b1e7ad32))
* **deps:** update module github.com/mattn/go-sqlite3 to v1.14.34 ([#85](https://github.com/meysam81/parse-dmarc/issues/85)) ([5336d3a](https://github.com/meysam81/parse-dmarc/commit/5336d3a64e647c32f46ddeb7d08c76fabd2a1200))
* **deps:** update module github.com/modelcontextprotocol/go-sdk to v1 ([#49](https://github.com/meysam81/parse-dmarc/issues/49)) ([58c0e79](https://github.com/meysam81/parse-dmarc/commit/58c0e79699392fe6661a8e6cce3092c4dd0d598f))
* **deps:** update module github.com/modelcontextprotocol/go-sdk to v1.3.1 ([#102](https://github.com/meysam81/parse-dmarc/issues/102)) ([68a80a4](https://github.com/meysam81/parse-dmarc/commit/68a80a4ab077af0ce4f45046211aaa77288de42b))
* **deps:** update module github.com/modelcontextprotocol/go-sdk to v1.4.0 ([#107](https://github.com/meysam81/parse-dmarc/issues/107)) ([39f1fcb](https://github.com/meysam81/parse-dmarc/commit/39f1fcb833098df1e2b3ab9d89c47350475055e6))
* **deps:** update module github.com/urfave/cli/v3 to v3.6.2 ([#25](https://github.com/meysam81/parse-dmarc/issues/25)) ([955e4e0](https://github.com/meysam81/parse-dmarc/commit/955e4e04d4d1d40617721ac4ec4aa70eaabad4ca))
* **deps:** update module github.com/urfave/cli/v3 to v3.7.0 ([#110](https://github.com/meysam81/parse-dmarc/issues/110)) ([8ce7090](https://github.com/meysam81/parse-dmarc/commit/8ce7090ce717ce25b563d2532b6a1a4abe7555b0))
* **deps:** update module modernc.org/sqlite to v1.45.0 ([#27](https://github.com/meysam81/parse-dmarc/issues/27)) ([2daa3b4](https://github.com/meysam81/parse-dmarc/commit/2daa3b4177bd5298e0f04eea8405fede27446ac2))
* **deps:** update module modernc.org/sqlite to v1.46.1 ([#101](https://github.com/meysam81/parse-dmarc/issues/101)) ([087c4e7](https://github.com/meysam81/parse-dmarc/commit/087c4e79f1fe433f82e2245a654c4e09701b69de))
* **docs:** update imap-host env var ([#116](https://github.com/meysam81/parse-dmarc/issues/116)) ([6d5aa4f](https://github.com/meysam81/parse-dmarc/commit/6d5aa4f32a13956723138ed786f599ad81b229c8))
* **goreleaser:** take build arg for copying binary in dockerfile ([5fb9504](https://github.com/meysam81/parse-dmarc/commit/5fb9504f1d7a946edf0c7c0a62eed13c1abbeb8d))

## [1.4.8](https://github.com/meysam81/parse-dmarc/compare/v1.4.7...v1.4.8) (2026-02-11)


### Bug Fixes

* **CI:** remove incomplete mcp docker from goreleaser ([356bb7b](https://github.com/meysam81/parse-dmarc/commit/356bb7b8ead6c3b783811178e49e10145a876d29))
* do not overwrite release notes from goreleaser ([43bec10](https://github.com/meysam81/parse-dmarc/commit/43bec1038b8a4ee46c6a898c89d918596bab0fdc))
* **mcp:** handle json error ([d17a04f](https://github.com/meysam81/parse-dmarc/commit/d17a04f23994911118f30522ea3ca909fb203efd))
* propagate errors with context ([85014c8](https://github.com/meysam81/parse-dmarc/commit/85014c8aafaf03a6d6b512a7be7eaccab971a4b0))
* stay compliant with brew core version print contraint ([35f8261](https://github.com/meysam81/parse-dmarc/commit/35f826169b3864811cc7684e7c87190518c373a5))

## [1.4.7](https://github.com/meysam81/parse-dmarc/compare/v1.4.6...v1.4.7) (2026-02-01)


### Bug Fixes

* **CI:** update mcp registry version identifier ([a135890](https://github.com/meysam81/parse-dmarc/commit/a135890747ef5fc07e43a7c9ed16e4dc21ac592b))

## [1.4.6](https://github.com/meysam81/parse-dmarc/compare/v1.4.5...v1.4.6) (2026-02-01)


### Bug Fixes

* make linter happy ([900e314](https://github.com/meysam81/parse-dmarc/commit/900e314e9c32909475db562c5b6d4b038496550e))
* **UI:** handle no data in compliance score ([#90](https://github.com/meysam81/parse-dmarc/issues/90)) ([eced435](https://github.com/meysam81/parse-dmarc/commit/eced435ab7b08a73829d4b7b6e09ed31d324834b))

## [1.4.5](https://github.com/meysam81/parse-dmarc/compare/v1.4.4...v1.4.5) (2026-01-25)


### Bug Fixes

* address unhandled charset ([#87](https://github.com/meysam81/parse-dmarc/issues/87)) ([99c245a](https://github.com/meysam81/parse-dmarc/commit/99c245a670ce1cb44f5d08b24b5d0c99c502ec69))

## [1.4.4](https://github.com/meysam81/parse-dmarc/compare/v1.4.3...v1.4.4) (2026-01-17)


### Bug Fixes

* **CI:** remove changelog from release note ([127cb80](https://github.com/meysam81/parse-dmarc/commit/127cb800f3f285b2d5899561fa390e2894bd52de))
* switch env and json config parsing order ([#83](https://github.com/meysam81/parse-dmarc/issues/83)) ([12afcd2](https://github.com/meysam81/parse-dmarc/commit/12afcd2e295c31b60a4696ba5c80edac1ae424e2)), closes [#84](https://github.com/meysam81/parse-dmarc/issues/84)

## [1.4.3](https://github.com/meysam81/parse-dmarc/compare/v1.4.2...v1.4.3) (2025-12-21)


### Features

* allow changing api endpoint from the settings modal ([#76](https://github.com/meysam81/parse-dmarc/issues/76)) ([e6cb48b](https://github.com/meysam81/parse-dmarc/commit/e6cb48be538e5c3e997d86dda9250068a2fb377e))


### Miscellaneous Chores

* prepare next release 1.4.3 ([141e2ed](https://github.com/meysam81/parse-dmarc/commit/141e2ed8c97a34e3d24d359645c0aa67237a47eb))

## [1.4.2](https://github.com/meysam81/parse-dmarc/compare/v1.4.1...v1.4.2) (2025-12-20)


### Bug Fixes

* **CI:** change mcp auth type to gh oidc ([1132b50](https://github.com/meysam81/parse-dmarc/commit/1132b5038c16fda2c83fe90a83f9f6e802c06af4))
* **dev:** minify html with vite plugin ([10eb4f7](https://github.com/meysam81/parse-dmarc/commit/10eb4f73ce936cc7def91d5aa633f2398b19fa33))

## [1.4.1](https://github.com/meysam81/parse-dmarc/compare/v1.4.0...v1.4.1) (2025-12-18)


### Features

* add DigitalOcean Droplet and Dokploy deployment options ([#67](https://github.com/meysam81/parse-dmarc/issues/67)) ([21d494f](https://github.com/meysam81/parse-dmarc/commit/21d494f60d7b7450d29cfd9a50fa25a914b100e8))
* **CI:** add mcp registry publishing to goreleaser ([a0d6f88](https://github.com/meysam81/parse-dmarc/commit/a0d6f88eb54df004c3b35ef7e256bb8d7be36cc9))


### Bug Fixes

* **docs:** add the missing png ([c9937ce](https://github.com/meysam81/parse-dmarc/commit/c9937cea0649a48ed92c973f223a25f8ef112f0f))
* use png for favicon ([4704ad7](https://github.com/meysam81/parse-dmarc/commit/4704ad7c23fbd31913adf677e50f001393a1adda))

## [1.4.0](https://github.com/meysam81/parse-dmarc/compare/v1.3.10...v1.4.0) (2025-12-17)


### Features

* **docs:** add 1-click deployment at the top of README ([#41](https://github.com/meysam81/parse-dmarc/issues/41)) ([50d2ff2](https://github.com/meysam81/parse-dmarc/commit/50d2ff2c54a506801e802daa6b5d375b79d1281d))
* **docs:** add northflank deployment button ([81215a5](https://github.com/meysam81/parse-dmarc/commit/81215a5b5e8cd87aa5f601f22fb460b4929d8d15))
* **docs:** add the self-hosted options to deployments ([42ab966](https://github.com/meysam81/parse-dmarc/commit/42ab966218b9e875ed438a9cb4a4b4ac3adb9b33))
* **docs:** add zeabur deployment template ([4abaa36](https://github.com/meysam81/parse-dmarc/commit/4abaa36d9b077f71a3e6da841af056dffbe046ba))
* **UI:** revamp the dashboard for actionable insights ([#68](https://github.com/meysam81/parse-dmarc/issues/68)) ([dbe073d](https://github.com/meysam81/parse-dmarc/commit/dbe073d35a7f349770e9fab6df4923e221c350f2))


### Bug Fixes

* **deps:** update module github.com/coreos/go-oidc/v3 to v3.17.0 ([#64](https://github.com/meysam81/parse-dmarc/issues/64)) ([306583c](https://github.com/meysam81/parse-dmarc/commit/306583cc37ba2c4fc7e7fffd542cdbab3edd119d))
* **docs:** clean up the providers ([6879ca7](https://github.com/meysam81/parse-dmarc/commit/6879ca70d2a9e5d1adc7f7732d2a85869aecdd2e))
* **docs:** reference dockerhub image ([b2d2b06](https://github.com/meysam81/parse-dmarc/commit/b2d2b0627dbf423d760b5ee9598060334b13f290))
* update docker image reference to dockerhub ([6afcb3a](https://github.com/meysam81/parse-dmarc/commit/6afcb3abe19beda674d19ab7911019c384ba54b1))

## [1.3.10](https://github.com/meysam81/parse-dmarc/compare/v1.3.9...v1.3.10) (2025-12-13)


### Bug Fixes

* **CI:** reverse the digest conditional for provenance ([9c0e165](https://github.com/meysam81/parse-dmarc/commit/9c0e165ec1cfeb58c7c535ec02bd0c81e60ba501))

## [1.3.9](https://github.com/meysam81/parse-dmarc/compare/v1.3.8...v1.3.9) (2025-12-13)


### Bug Fixes

* **CI:** perform keyless docker sign ([1f322d3](https://github.com/meysam81/parse-dmarc/commit/1f322d33999d3586880d82c5ce3570b60157cd21))
* **CI:** update the digest name for attestation ([eb692a8](https://github.com/meysam81/parse-dmarc/commit/eb692a81fb08205990e639ab6c54c8998c705cd0))

## [1.3.8](https://github.com/meysam81/parse-dmarc/compare/v1.3.7...v1.3.8) (2025-12-13)


### Bug Fixes

* **CI:** set up buildx action for multi platform build ([a36e41c](https://github.com/meysam81/parse-dmarc/commit/a36e41c02e6af54297275042e269c66b85545a78))

## [1.3.7](https://github.com/meysam81/parse-dmarc/compare/v1.3.6...v1.3.7) (2025-12-13)


### Bug Fixes

* **CI:** remove annotations from docker build ([b5d0cd6](https://github.com/meysam81/parse-dmarc/commit/b5d0cd650ba0dd780ea95ec35c82940620aad6bb))

## [1.3.6](https://github.com/meysam81/parse-dmarc/compare/v1.3.5...v1.3.6) (2025-12-13)


### Bug Fixes

* **CI:** disable sbom via config ([6b0f3be](https://github.com/meysam81/parse-dmarc/commit/6b0f3be7d79cafe1eeb0930d2f30bf02f53c56af))

## [1.3.5](https://github.com/meysam81/parse-dmarc/compare/v1.3.4...v1.3.5) (2025-12-13)


### Bug Fixes

* **CI:** disable attestation digest for docker build ([95d6665](https://github.com/meysam81/parse-dmarc/commit/95d666543f59b3a890ea8921275f840eebb32f8e))

## [1.3.4](https://github.com/meysam81/parse-dmarc/compare/v1.3.3...v1.3.4) (2025-12-13)


### Bug Fixes

* **CI:** disable sbom on docker ([36e5cc4](https://github.com/meysam81/parse-dmarc/commit/36e5cc4d829efa633fc75879e0a132d076866318))

## [1.3.3](https://github.com/meysam81/parse-dmarc/compare/v1.3.2...v1.3.3) (2025-12-13)


### Bug Fixes

* **CI:** add current dir to docker build context ([254cc6a](https://github.com/meysam81/parse-dmarc/commit/254cc6a6b0b88afc67af8a1cdc4a41d1b71ff1a6))

## [1.3.2](https://github.com/meysam81/parse-dmarc/compare/v1.3.1...v1.3.2) (2025-12-13)


### Features

* **CI:** build docker via goreleaser ([8555265](https://github.com/meysam81/parse-dmarc/commit/855526505aa48c9ab1cdc946f10370a875ee47d8))

## [1.3.1](https://github.com/meysam81/parse-dmarc/compare/v1.3.0...v1.3.1) (2025-12-13)


### Bug Fixes

* **CI:** publish to casks for a change ([0e8346b](https://github.com/meysam81/parse-dmarc/commit/0e8346b3e1c3dcc557445370f59ea7c93588c1fd))

## [1.3.0](https://github.com/meysam81/parse-dmarc/compare/v1.2.2...v1.3.0) (2025-12-12)


### Features

* add MCP server integration to project ([#46](https://github.com/meysam81/parse-dmarc/issues/46)) ([807b8d6](https://github.com/meysam81/parse-dmarc/commit/807b8d677c81aef6ac39605bffb7125680841d4a))
* **CI:** add man page and shell completion to brew ([0e45ed2](https://github.com/meysam81/parse-dmarc/commit/0e45ed2a10d26a06fadc6a2fdfe3b81ad5a01197))
* **CI:** add prettier job ([39af52d](https://github.com/meysam81/parse-dmarc/commit/39af52d186fcf75dba2f8ac05ef313808f1e01b4))
* **CI:** install pandoc and use official goreleaser action ([0d0a307](https://github.com/meysam81/parse-dmarc/commit/0d0a3077395a9879a9d036166b4afbab43ee164e))
* **mcp:** add OAuth2 authentication for MCP HTTP server ([361f078](https://github.com/meysam81/parse-dmarc/commit/361f0782e2f20b7160873a085e600adc68988432))

## [1.2.2](https://github.com/meysam81/parse-dmarc/compare/v1.2.1...v1.2.2) (2025-12-05)


### Bug Fixes

* **CI:** add the brew token ([4081621](https://github.com/meysam81/parse-dmarc/commit/4081621ae9d1771185712512a912aa350a2e0305))

## [1.2.1](https://github.com/meysam81/parse-dmarc/compare/v1.2.0...v1.2.1) (2025-12-05)


### Bug Fixes

* **CI:** move bundled dist to server directory ([59dabf1](https://github.com/meysam81/parse-dmarc/commit/59dabf1becfc22fddda4bfdab6dd22188396b4c3))
* **docs:** pin to full version for now ([2060e45](https://github.com/meysam81/parse-dmarc/commit/2060e451a3b7322d933ad4b8f389f28474f85fb1))
* **docs:** update quickstart command with volume ([617c258](https://github.com/meysam81/parse-dmarc/commit/617c2584402cb309e1a652630f5f75de92df0b2e))

## [1.2.0](https://github.com/meysam81/parse-dmarc/compare/v1.1.0...v1.2.0) (2025-12-05)


### Features

* add production-ready Grafana dashboard ([6f857fe](https://github.com/meysam81/parse-dmarc/commit/6f857fe660144093b8766f529c15b17a21edc3b2))
* add production-ready Prometheus metrics ([#39](https://github.com/meysam81/parse-dmarc/issues/39)) ([17f6968](https://github.com/meysam81/parse-dmarc/commit/17f69681b23eee92ff1bc908fde7da2de11b8a52))
* **frontend:** add DMARC DNS record generator ([34491ba](https://github.com/meysam81/parse-dmarc/commit/34491ba244f9af48a25e3a0d6b09979661678c0d))
* **frontend:** implement dark mode with theme toggle ([#35](https://github.com/meysam81/parse-dmarc/issues/35)) ([569fd36](https://github.com/meysam81/parse-dmarc/commit/569fd3649931f94fa4f39845d0f141afacd6fbc0))


### Bug Fixes

* **CI:** build on major version as well ([fdff51d](https://github.com/meysam81/parse-dmarc/commit/fdff51d2c77d55e890e5b6d9c12ce316ddec2542))
* **CI:** only pin main with latest ([f120b2e](https://github.com/meysam81/parse-dmarc/commit/f120b2e7b7985adbaa815b6104f8aaee44e8e9ce))
* **CI:** reverse the conditional for build-dev job ([ce8b417](https://github.com/meysam81/parse-dmarc/commit/ce8b417c65c62563886787df3871200679c8de66))
* **CI:** update dockerignore after main.go change ([1241b2f](https://github.com/meysam81/parse-dmarc/commit/1241b2f22b8138348b7bf087ba00659e0b1f6fe8))
* **CI:** update version of the binary ([0a8f023](https://github.com/meysam81/parse-dmarc/commit/0a8f0230b30229349db270aa065e3332da63990a))
* **CI:** use go version file when setting up go ([40fbcb1](https://github.com/meysam81/parse-dmarc/commit/40fbcb1dfbe962c7d48ef2c821ee32d997e7c78c))
* **dev:** remove extra compose file ([eb71b09](https://github.com/meysam81/parse-dmarc/commit/eb71b097005bb9d024454ee6d84c2ce5fbb59d7b))
* **dev:** remove user from compose ([e9fbf73](https://github.com/meysam81/parse-dmarc/commit/e9fbf73b72a2a9958d62ed6f6ffea4bad2f2771f))
* make github star non-intrusive ([831047b](https://github.com/meysam81/parse-dmarc/commit/831047b88394cd468ff57bc08c53907880e0feee))
* move main.go to the root ([0dd5787](https://github.com/meysam81/parse-dmarc/commit/0dd57878ec5f7d6707faca0189834a6d91c266e5))
* publish to grafana community dashboards instead ([86f40b0](https://github.com/meysam81/parse-dmarc/commit/86f40b052607861de3d1a843f9e9c9e08ba57481))
* remove broken dns generator ([df39352](https://github.com/meysam81/parse-dmarc/commit/df393527710b487bc0bdbe7d67d9a3f909623b92))

## [1.1.0](https://github.com/meysam81/parse-dmarc/compare/v1.0.2...v1.1.0) (2025-11-08)


### Features

* **docs:** add nerdy badges to README ([#23](https://github.com/meysam81/parse-dmarc/issues/23)) ([5026a45](https://github.com/meysam81/parse-dmarc/commit/5026a45da22a9fa0a3c831945da4afaa05c2dd7c))


### Bug Fixes

* **CI:** update goreleaser after moving FE to root ([357d37e](https://github.com/meysam81/parse-dmarc/commit/357d37ed69d9cd7400c05c2f7ca2667072f4d46d))

## [1.0.2](https://github.com/meysam81/parse-dmarc/compare/v1.0.1...v1.0.2) (2025-11-07)


### Bug Fixes

* **dev:** bring the frontend to the root and simplify docker ([edc64e7](https://github.com/meysam81/parse-dmarc/commit/edc64e7417c910b7200bb1d727f0ba200c1a787d))

## [1.0.1](https://github.com/meysam81/parse-dmarc/compare/v1.0.0...v1.0.1) (2025-11-07)


### Bug Fixes

* **docs:** simplify README for first-time viewers ([#15](https://github.com/meysam81/parse-dmarc/issues/15)) ([7341b7b](https://github.com/meysam81/parse-dmarc/commit/7341b7bb022e5e96798e6d8d3523a377e282cf6f))

## 1.0.0 (2025-11-05)


### Features

* add CI and goreleaser with optimized distroless docker ([66c6828](https://github.com/meysam81/parse-dmarc/commit/66c682807d5ed12328349e2713c9d814d6337e78))
* add sort and refresh to UI ([9d6853e](https://github.com/meysam81/parse-dmarc/commit/9d6853e95319529c40958ae60d95d05d8cb5a675))
* clean up the cli and optimize the docker ([db23ba4](https://github.com/meysam81/parse-dmarc/commit/db23ba4ff2fd80ad1c016daa73a57519998dd271))
* **docs:** update readme to the latest changes ([c20e0cf](https://github.com/meysam81/parse-dmarc/commit/c20e0cf837b91b05e61153e4e55e7315fb09d396))
* ensure db path exists and add demo screenshot ([42d6d14](https://github.com/meysam81/parse-dmarc/commit/42d6d14fd96f2818830c73ad4e23b8f97719c497))
* Implement DMARC report parser with Vue.js dashboard ([#1](https://github.com/meysam81/parse-dmarc/issues/1)) ([23b8ac0](https://github.com/meysam81/parse-dmarc/commit/23b8ac0cc63a81aad7dfa2385ed24f6d74915775))
* update the UI footer with OS friendly text ([5008fdf](https://github.com/meysam81/parse-dmarc/commit/5008fdf5e59e5b4af1fc12ccafc644d27ccdbdfc))
* update UI footer & readme and optimize vite ([9578d05](https://github.com/meysam81/parse-dmarc/commit/9578d0530adab10334ad1a8a20c563abdc589854))


### Bug Fixes

* **CI:** ensure a fake dist exists before golangci-lint ([4e01983](https://github.com/meysam81/parse-dmarc/commit/4e01983f13a16ab0a8ea7cb880e43b39c285ba1f))
* **CI:** make linter happy ([cd04dfc](https://github.com/meysam81/parse-dmarc/commit/cd04dfc8beae30b0405d38834680011cd7d14fe5))
* make server context aware and do not fail UI on empty result ([588c466](https://github.com/meysam81/parse-dmarc/commit/588c4660534d4111cf24738a952cbad56ebbf345))
* update the UI on reports API parsing ([b85b59e](https://github.com/meysam81/parse-dmarc/commit/b85b59ef577ffeb6b4dacd1188e6e55d549b1b5a))
* use app for static config ([3820862](https://github.com/meysam81/parse-dmarc/commit/3820862f8fffe26fd3dad7c004d1c0b04fc8be9e))
