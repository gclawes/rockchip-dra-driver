# [0.7.0](https://github.com/gclawes/rockchip-dra-driver/compare/0.6.0...0.7.0) (2026-09-29)


### Features

* **rga:** advertise rockchip-rga ([c8f8131](https://github.com/gclawes/rockchip-dra-driver/commit/c8f8131639e1572d94acfc7738cc0961b3aab4d0))
* **vpu:** advertise mainline decoder and encoder nodes ([3aac1c8](https://github.com/gclawes/rockchip-dra-driver/commit/3aac1c8d66efd64b9e6d821d4d97846be2f81e19))

# [0.6.0](https://github.com/gclawes/rockchip-dra-driver/compare/0.5.0...0.6.0) (2026-09-29)


### Bug Fixes

* **deps:** bump the kubernetes group with 5 updates ([ffbee76](https://github.com/gclawes/rockchip-dra-driver/commit/ffbee76d961133c6e24afbf4a534ed52a7ce0243))


### Features

* constrain share requests ([54c3f56](https://github.com/gclawes/rockchip-dra-driver/commit/54c3f5695e775a92822d81abd33a925064ed9620))

# [0.5.0](https://github.com/gclawes/rockchip-dra-driver/compare/0.4.1...0.5.0) (2026-09-29)


### Features

* **cdi:** publish device node gid ([06b5502](https://github.com/gclawes/rockchip-dra-driver/commit/06b5502357c9ebd5978d8e93e836b7d8d9ce7207))

## [0.4.1](https://github.com/gclawes/rockchip-dra-driver/compare/0.4.0...0.4.1) (2026-09-28)


### Bug Fixes

* push arm images with buildx attestations ([1c0ef22](https://github.com/gclawes/rockchip-dra-driver/commit/1c0ef22ab37b4c97531b58ff3bfff98433347edc))

# [0.4.0](https://github.com/gclawes/rockchip-dra-driver/compare/0.3.0...0.4.0) (2026-09-25)


### Features

* **discovery:** report health and rediscover devices ([fb9815e](https://github.com/gclawes/rockchip-dra-driver/commit/fb9815e6a25fefd05b5625ea2b47f1c6ac8d9702))

# [0.3.0](https://github.com/gclawes/rockchip-dra-driver/compare/0.2.0...0.3.0) (2026-09-13)


### Features

* use distroless static debian13 as the runtime image ([b509ad7](https://github.com/gclawes/rockchip-dra-driver/commit/b509ad7f3bbd7907423bd56a524b3840e4b5e7cd))

# [0.2.0](https://github.com/gclawes/rockchip-dra-driver/compare/0.1.5...0.2.0) (2026-09-11)


### Features

* scope CDI env vars by device type ([06bba6b](https://github.com/gclawes/rockchip-dra-driver/commit/06bba6b80081c6a8956da0d5e30d9f8755912f27))

## [0.1.5](https://github.com/gclawes/rockchip-dra-driver/compare/0.1.4...0.1.5) (2026-09-06)


### Bug Fixes

* **helm:** allow install on kubernetes 1.35 ([754d67b](https://github.com/gclawes/rockchip-dra-driver/commit/754d67b2de3099fe262ee8d7f2e372e00a6c0f6e))

## [0.1.4](https://github.com/gclawes/rockchip-dra-driver/compare/0.1.3...0.1.4) (2026-09-06)


### Bug Fixes

* **discovery:** detect rocket via bound NPU cores ([4e0e62c](https://github.com/gclawes/rockchip-dra-driver/commit/4e0e62cf147c5ac7d050dcbc207b201c179c5be8))

## [0.1.3](https://github.com/gclawes/rockchip-dra-driver/compare/0.1.2...0.1.3) (2026-09-06)


### Bug Fixes

* use specs-go MinimumRequiredVersion for CDI ([1378974](https://github.com/gclawes/rockchip-dra-driver/commit/137897482fb4d2318ad6d49186c2c7551c277da8))

## [0.1.2](https://github.com/gclawes/rockchip-dra-driver/compare/0.1.1...0.1.2) (2026-09-06)


### Bug Fixes

* **ci:** use kind 0.33 and disable govet inline ([3f16516](https://github.com/gclawes/rockchip-dra-driver/commit/3f16516acc9d53013cbca848c1d5661e6b8ae545))

## [0.1.1](https://github.com/gclawes/rockchip-dra-driver/compare/0.1.0...0.1.1) (2026-09-06)


### Bug Fixes

* unblock lint and e2e setup ([5f7e41b](https://github.com/gclawes/rockchip-dra-driver/commit/5f7e41b27bae11ccc79d6a3a65af2e4cb5d36c4f))

# [0.1.0](https://github.com/gclawes/rockchip-dra-driver/compare/0.0.0...0.1.0) (2026-09-06)


### Features

* add kubelet plugin that publishes shared NPU and GPU ([ea8e9c1](https://github.com/gclawes/rockchip-dra-driver/commit/ea8e9c1d6da9d82d53620bac0803ae65c5d3df33))
* **api:** add v1alpha1 NPU and GPU opaque config types ([d5e9b82](https://github.com/gclawes/rockchip-dra-driver/commit/d5e9b82fc856d3f3b93995713ebc0d7844912f6e))
* **discovery:** enumerate rocket NPU and panthor GPU ([a65942b](https://github.com/gclawes/rockchip-dra-driver/commit/a65942bef7d587a588cda0b256244041e49ccdba))
* **helm:** add chart, DeviceClasses, and allocation caps ([79383c0](https://github.com/gclawes/rockchip-dra-driver/commit/79383c08196d294701e3ff39b0bcda11e33e4eeb))

# Changelog

All notable changes to this project will be documented in this file.
