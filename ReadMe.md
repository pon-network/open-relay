# Welcome To PoN Open Relay

The Proof of Neutrality Relay (PoN) Open relay is a permissionless and neutral header-only relay in optimistic mode with retrospective support for Flashbot builders in open relay mode where the transaction contents are thrown away at source and not used for block simulation, offering a hydbrid model to migrate to fully header-only mode when credible commitment proofs are used to slash misbehaving builders as per the relay rule.

Key characteristics of Open Relay

- Use onchain builder account (pon-builder registry contract)
- Use onchain proposer account ( pon - proposer registry contract)
- Unified builder collateral for all open relay instances
- PoN reporter (ZK gadget) slash any commitment deviation onchain

[![Relay Documentation](https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667)](https://docs.pon.network/pon/relay)
[![Relay Documentation](https://img.shields.io/badge/Documentation-Docusaurus-green)](https://docs.pon.network/)

### Easy Install

Install latest release from [pon-network](https://github.com/pon-network/open-relay/releases)

Install and run using-

```shell
$ go install github.com/pon-network/open-relay
$ open-relay --help
```

## Setting Up The Relay and Building

### Building From Binaries

Automated builds are available for stable releases and the unstable master branch. Binary
archives are published at https://github.com/pon-network/open-relay.

Use the binaries along with the [![Relay Documentation](https://img.shields.io/badge/Documentation-Docusaurus-green)](https://docs.pon.network/) to install and run relay from binaries

### Building From Source

To run from source use the main branch of this repository and use the following command-

```shell
go migrate
go run . relay \
--relay-url <Relay_URL> \
--beacon-uris <Beacon_URIS> \
--redis-uri <Redis_URIS> \
--db <DB_URL> \
--secret-key <Relay_BLS> \
--network <Network> \
--max-db-connections <Max_DB_Connections> \
--max-idle-connections <Max_Idle_Connections> \
--max-idle-timeout <Max_Idle_Timeout> \
--db-driver <DB_Driver> \
--pon-pool <PON_POOL_URL> \
--pon-pool-API-Key <PON_POOL_API_KEY> \
--bulletinBoard-broker <Bulletin_Board_Broker> \
--bulletinBoard-port <Bulletin_Board_Port> \
--bulletinBoard-client <Bulletin_Board_Client> \
--bulletinBoard-password <Bulletin_Board_Password> \
--reporter-url <Reporter_URL> \
--bid-timeout <Bid_Timeout> \
--relay-read-timeout <Relay_Read_Timeout> \
--relay-read-header-timeout <Relay_Read_Header_Timeout>
--relay-write-timeout <Relay_Write_Timeout> \
--relay-idle-timeout <Relay_Idle_Timeout> \
--open-relay
```

#### Relay Services

![](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![](https://img.shields.io/badge/redis-%23DD0031.svg?&style=for-the-badge&logo=redis&logoColor=white)

If You Want To Run A Postgres Server And Redis Use The Provided Docker Compose File By Following The [Docker Documentation](./docker-compose/Readme.md)
` If you use your own POSTGRES_DB_USER and POSTGRES_DB_PASSWORD, please update parameter when running relay based on that`

#### Metabase

![](https://img.shields.io/badge/Metabase-509EE3?style=for-the-badge&logo=metabase&logoColor=fff)

PON Relay comes with a metabase docker-compose file. It can be used by following the [Docker Documentation](./docker-compose/Readme.md).
` For Performance Its Better If You Run The Service In Different Machine Then The Relay`

#### PoN Builder Parameters

| Parameter                     | Description                                                              | Default            | Required |
| ----------------------------- | ------------------------------------------------------------------------ | ------------------ | -------- |
| `--relay-url`                 | Listen Address For The PoN Relay Service Locally                         | `"localhost:9000"` | No       |
| `--beacon-uris`               | Beacon Node Endpoint                                                     | `""`               | Yes      |
| `--db`                        | Database URL                                                             | `""`               | Yes      |
| `--secret-key`                | BLS Secret Key Of Relay                                                  | `""`               | Yes      |
| `--network`                   | Network `(Testnet/ Mainnet)`                                             | `"Testnet"`        | No       |
| `--max-db-connections`        | Maximum Database Connections                                             | `100`              | No       |
| `--max-idle-connections`      | Maximum Database Idle Connections                                        | `100`              | No       |
| `--max-idle-timeout`          | Maximum Database Timeout `(In 1s/ 5h format)`                            | `100s`             | No       |
| `--db-driver`                 | Database Driver                                                          | `postgres`         | No       |
| `--pon-pool`                  | Pon Pool Subgraph URL                                                    | `""`               | Yes      |
| `--bulletinBoard-broker`      | Bulletin Board MQTT Broker URL                                           | `""`               | Yes      |
| `--bulletinBoard-port`        | Bulletin Board MQTT Port                                                 | `""`               | Yes      |
| `--bulletinBoard-client`      | Bulletin Board Client                                                    | `""`               | Yes      | ̦    |
| `--bulletinBoard-password`    | Bulletin Board Password                                                  | `""`               | Yes      |
| `--bid-timeout`               | Maximum Time Bid Is Kept With Relay `(In 1s/ 5h format)`                 | `"15s"`            | No       |
| `--relay-read-timeout`        | Relay Server Read Timeout `(In 1s/ 5h format)`                           | `"10s"`            | No       |
| `--relay-read-header-timeout` | Relay Server Read Header Timeout `(In 1s/ 5h format)`                    | `"10s"`            | No       |
| `--relay-write-timeout`       | Relay Server Write Timeout `(In 1s/ 5h format)`                          | `"10s"`            | No       |
| `--relay-idle-timeout`        | Relay Idle Timeout `(In 1s/ 5h format)`                                  | `"10s"`            | No       |
| `--new-relic-application`     | New Relic Application `(New Relic Not Used If Application Not Provided)` | `""`               | No       |
| `--new-relic-license`         | New Relic License                                                        | `""`               | No       |
| `--new-relic-forwarding`      | New Relic Forwarding                                                     | `false`            | No       |
| `--open-relay`                | Run Open Relay                                                           | `false`            | Yes      |

## DB Differences Between Open Relay and Proof Relay

Open Relay needs to maintain the validators in beacon chain and hence open relay takes more space to run for database than proof relay. The different database setup is handled by the relay migrations only based on whether the relayy is run with --openRelay.

## Hardware Requirements

![](https://img.shields.io/badge/Coming-Soon-red)

## API Spec

[![API Spec](https://validator.swagger.io/validator/?url=https%3A%2F%2Fgithub.com%2Fbsn-eng%2FPon-relay%2FDocs%2Fswagger.yaml)](./docs/swagger.yaml)

API Spec Is Generated Using Go-Swagger for Open API Implementation. You can follow the specification by visiting the [API-Spec](./docs/APISpec.md)
If you want to make changes to the swagger file, run the following commands-

```shell
make swagger
make serve-swagger
```
