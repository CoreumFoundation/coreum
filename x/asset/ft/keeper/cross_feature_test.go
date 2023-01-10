package keeper

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
// # All in one
// - Core + FT coin in one send and multi-send with All Features enabled and account state which already include the whitelisted and frozen state
//
//# Permission matrix
// - Add table test that check the access to all features by issuer and non issuer with the enabled and disabled features
