package keeper

// # Issuer + Whitelist
// - whitelisting balance for issuer account should not be possible, issuer may always receive and send funds
//   so whitelisting him may create a lot of troubles
// - limiting issuer does not provide any value because he may always remove/change any limits
// - tx whitelisting balance for issuer account should error out
//
// # Issuer + Freeze
// - freezing balance for issuer account should not be possible, issuer may always receive and send funds
//   so freezing him may create a lot of troubles
// - limiting issuer does not provide any value because he may always remove/change any limits
// - tx freezing balance for issuer account should error out
//
// # Issuer + Global Freeze
// - on the other hand global freeze means that nobody can send and receive tokens
// - if issuer sends/receives tokens from/to someone else it is enforced because the other account is not an issuer
// - but it should apply even if issuer sends tokens to himself
// - in general if global freeze is active all transfers must be blocked, including those defined by the issuer
//
// # Mint + Global Freeze
// - if global freeze is active minting should fail
//
// # Burn + Global Freeze
// - if global freeze is active burning should fail
//
// # Mint + Burn rate
// - burn rate should not apply when tokens are received by the issuer because of minting
//
// # Mint + Send commission
// - send commission should not apply when tokens are received by the issuer because of minting
//
// # Burn + Burn rate
// - burn rate should not apply when tokens are burnt
//
// # Burn + Send commission
// - send commission should not apply when tokens are burnt
//
// # Burn + Freeze
// - when someone wants to burn more than is allowed by non-frozen balance tx should fail
//
// # Burn rate + Freeze
// - when burn rate causes a non-frozen balance to be exceeded, tx should fail
//
// # Send commission + Freeze
// - when send commission causes a non-frozen balance to be exceeded, tx should fail
