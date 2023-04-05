# Messages types

The following are the supported messages types that are exchanged between the client and the server through the websocket.

| Type                    | Description                                                                                                           |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `ERROR`                 | Unexpected server side error. You can find more information in the message description / payload                      |
| `GYM_ATTACK_CANDIDATE`  | It announces an incoming attack. The user has the opportunity to dodge it using the `GYM_ATTACK_DODGED` message type. |
| `GYM_ATTACK_DODGED`     | The user avoids the gym Loomie attack. Has a 1 second cooldown                                                        |
| `USER_LOOMIE_WEAKENED`  | The user Loomie was defeated by the gym Loomie                                                                        |
| `UPDATE_PLAYER_LOOMIE`  | The current user Loomie was changed                                                                                   |
| `UPDATE_USER_LOOMIE_HP` | The current user Loomie was attacked by the gym Loomie                                                                |
| `USER_HAS_LOST`         | All the user Loomies were defeated                                                                                    |
| `USER_ATTACK`           | It will reduce the gym Loomie hp. The enemy Loomie has a chance to dodge it (10%). Has a 1 second cooldown            |
| `USER_ATTACK_DODGED`    | The gym avois the user Loomie attack                                                                                  |
| `GYM_LOOMIE_WEAKENED`   | The gym Loomie was defeated by the user Loomie                                                                        |
| `UPDATE_GYM_LOOMIE`     | The current gym Loomie was changed                                                                                    |
| `UPDATE_GYM_LOOMIE_HP`  | The current gym Loomie was attacked by the user Loomie                                                                |
| `USER_HAS_WON`          | All the gym Loomies were defeated                                                                                     |
