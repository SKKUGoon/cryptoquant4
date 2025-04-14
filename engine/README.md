# Redis Requirements

Redis cache should be:

- **Fast**: In-memory, ephemeral storage
- **Safe**: Prevent fund overspending and race conditions 
- **Shared**: Accessible across all containers
- **Minimal**: Store only essential real-time state

# On Startup (cryptoquant_init)

1. Fetch Wallet
2. Set available fund (fetch from database)
3. Zero reserved fund
4. Set wallet snapshot

# On Order Completion / Failure
1. Market Order - the order is always going to be filled 
2. Failure - No failure
3. Update the reserve fund

# On Position Close
1. Remove the position
2. Increase the available fund