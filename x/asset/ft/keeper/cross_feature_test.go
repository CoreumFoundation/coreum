package keeper

// # All tests must be applied for to send and multi-send

// # Issue bare toke
// - issue the FT with the minimum possible filled data and apply Mint/Freeze/Whitelist/Burn

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
// # Burn rate + Send Commission + Freeze
// - when there is enough unfrozen coins for the burn rate, but not enough for commission (and vice versa), tx should fail
//
// # Send commission + Freeze
// - when send commission causes a non-frozen balance to be exceeded, tx should fail
//
// # Sending cored + FT
// - test sending FT together with cored when FT transfer should fail/succeed due to whitelist/freezing
//
// # Balance + Burn rate
// - burn rate should not apply when there are not enough tokes to burn and send
//
// # Balance + Send rate
// - send rate should not apply when there are not enough tokes for send rate and send
//
// $ Non-issuer + Whitelist (add this test directly to whitelisting if there are no such)
// - whitelist non-issuer to have 10 coins, send 10 coins to non-issuer, try to set non-issuer whitelist amount to 9, the tx should failed
//
// # All in one
// - Core + FT coin in one send and multi-send with All Features enabled and account state which already include the whitelisted and frozen state
//
//# Permission matrix
// - Add table test that check the access to all features by issuer and non issuer with the enabled and disabled features
//
//# Native tokens check
// - Add table test to check that a native token skips all features checks
//
//# Authz
// - Add test to issue a token with freeze and whitelisting features and test those features from the issuer and non-issuer accounts executed on behalf of a grantee
//
