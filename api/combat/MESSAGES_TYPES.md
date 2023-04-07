# Messages types

The following are the supported messages types that are exchanged between the client and the server through the websocket.

| Type                    | Description                                                                                                           | From   | To     |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------- | ------ | ------ |
| `ERROR`                 | Unexpected server side error. You can find more information in the message description / payload                      | Server | Client |
| `GYM_ATTACK_CANDIDATE`  | It announces an incoming attack. The user has the opportunity to dodge it using the `GYM_ATTACK_DODGED` message type. | Server | Client |
| `USER_DODGE`            | The user avoids the gym Loomie attack. Has a 1 second cooldown                                                        | Client | Server |
| `GYM_ATTACK_DODGED`     | Confirmation that the user avoids the gym Loomie attack                                                               | Server | Client |
| `USER_LOOMIE_WEAKENED`  | The user Loomie was defeated by the gym Loomie                                                                        | Server | Client |
| `UPDATE_PLAYER_LOOMIE`  | The current user Loomie was changed                                                                                   | Server | Client |
| `UPDATE_USER_LOOMIE_HP` | The current user Loomie was attacked by the gym Loomie                                                                | Server | Client |
| `USER_HAS_LOST`         | All the user Loomies were defeated                                                                                    | Server | Client |
| `USER_ATTACK`           | It will reduce the gym Loomie hp. The enemy Loomie has a chance to dodge it (10%). Has a 1 second cooldown            | Client | Server |
| `USER_ATTACK_DODGED`    | The gym avois the user Loomie attack                                                                                  | Server | Client |
| `GYM_LOOMIE_WEAKENED`   | The gym Loomie was defeated by the user Loomie                                                                        | Server | Client |
| `UPDATE_GYM_LOOMIE`     | The current gym Loomie was changed                                                                                    | Server | Client |
| `UPDATE_GYM_LOOMIE_HP`  | The current gym Loomie was attacked by the user Loomie                                                                | Server | Client |
| `USER_HAS_WON`          | All the gym Loomies were defeated                                                                                     | Server | Client |