## Combat Sequence Diagram

The following diagram shows the sequence of events that occur when a user starts a combat.

```mermaid
sequenceDiagram
  Client ->> /combat/register: Request combat token
  /combat/register ->> /combat/register: Validate the access token

  /combat/register -->> Client: Combat token
  Client ->> /combat: Request protocol upgrade
  /combat ->> /combat: Validate the combat token
   /combat ->> /combat: Check the gym is not already in combat
  /combat ->> /combat: Check the user is near the gym
  /combat ->> /combat: Check the user and gyn have at least one loomie
  /combat -->> Client: Upgrade protocol to WS
  Ws Connection ->> Client: Send the initial combat state
  Client ->> Ws Connection: Exchange combat actions
  Ws Connection ->> Ws Connection: Validate the combat actions
  Ws Connection ->> Ws Connection: Apply the combat actions
  Ws Connection ->> Client: Exchange the combat actions
```
