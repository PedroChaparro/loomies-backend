## Messages types

The following are the supported messages types that are exchanged between the client and the server through the websocket.

| Type                     | Description                                                                                                           | From   | To     |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------- | ------ | ------ |
| `ERROR`                  | Unexpected server side error. You can find more information in the response message / payload                         | Server | Client |
| `ERROR_USING_ITEM`       | The user can't use the item. You can find more information in the response message / payload                          | Server | Client |
| `COMBAT_TIMEOUT`         | The combat has been closed due to the player inactivity.                                                              | Server | Client |
| `USER_USE_ITEM`          | The user uses an item in the combat.                                                                                  | CLient | Server |
| `USER_ITEM_USED`         | Confirmation that the user uses an item in the combat.                                                                | Server | Client |
| `USER_CHANGE_LOOMIE`     | The user changes the current Loomie.                                                                                  | Client | Server |
| `GYM_ATTACK_CANDIDATE`   | It announces an incoming attack. The user has the opportunity to dodge it using the `GYM_ATTACK_DODGED` message type. | Server | Client |
| `USER_DODGE`             | The user avoids the gym Loomie attack. Has a 1 second cooldown                                                        | Client | Server |
| `GYM_ATTACK_DODGED`      | Confirmation that the user avoids the gym Loomie attack                                                               | Server | Client |
| `USER_LOOMIE_WEAKENED`   | The user Loomie was defeated by the gym Loomie                                                                        | Server | Client |
| `UPDATE_USER_LOOMIE`     | The current user Loomie was changed                                                                                   | Server | Client |
| `UPDATE_USER_LOOMIE_HP`  | The current user Loomie was attacked by the gym Loomie                                                                | Server | Client |
| `UPDATE_USER_LOOMIE_EXP` | The current user Loomie experience was updated                                                                        | Server | Client |
| `USER_HAS_LOST`          | All the user Loomies were defeated                                                                                    | Server | Client |
| `USER_ESCAPE_COMBAT`     | The user escapes from combat                                                                                          | Client | Server |
| `ESCAPE_COMBAT`          | Message when the user escapes combat                                                                                  | Server | Client |
| `USER_ATTACK`            | It will reduce the gym Loomie hp. The enemy Loomie has a chance to dodge it (10%). Has a 1 second cooldown            | Client | Server |
| `USER_ATTACK_DODGED`     | The gym avoid the user Loomie attack                                                                                  | Server | Client |
| `GYM_LOOMIE_WEAKENED`    | The gym Loomie was defeated by the user Loomie                                                                        | Server | Client |
| `UPDATE_GYM_LOOMIE`      | The current gym Loomie was changed                                                                                    | Server | Client |
| `UPDATE_GYM_LOOMIE_HP`   | The current gym Loomie was attacked by the user Loomie                                                                | Server | Client |
| `USER_HAS_WON`           | All the gym Loomies were defeated                                                                                     | Server | Client |
| `USER_GET_LOOMIE_TEAM`   | Get loomies team from user                                                                                            | Client | Server |
| `USER_LOOMIE_TEAM`       | Loomies team response                                                                                                 | Server | Client |

## Payloads

The following are the payloads that are sent with some of the messages types.

### USER_USE_ITEM

The application must send the item id to the server as a payload.

```json
{
  "type": "USER_USE_ITEM",
  "payload": {
    "item_id": "The mongo id of the item"
  }
}
```

### USER_CHANGE_LOOMIE

The application must send the Loomie id to the server as a payload.

```json
{
  "type": "USER_CHANGE_LOOMIE",
  "payload": {
    "loomie_id": "The mongo id of the Loomie"
  }
}
```
