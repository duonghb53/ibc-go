package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

func EmitOnRecvPacketEvents(ctx sdk.Context, packetData types.FungibleTokenPacketDataV2, ack channeltypes.Acknowledgement, ackErr error) {
	eventAttributes := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeKeySender, packetData.Sender),
		sdk.NewAttribute(types.AttributeKeyReceiver, packetData.Receiver),
		sdk.NewAttribute(types.AttributeKeyTokens, types.Tokens(packetData.Tokens).String()),
		sdk.NewAttribute(types.AttributeKeyMemo, packetData.Memo),
		sdk.NewAttribute(types.AttributeKeyAckSuccess, fmt.Sprintf("%t", ack.Success())),
	}

	if ackErr != nil {
		eventAttributes = append(eventAttributes, sdk.NewAttribute(types.AttributeKeyAckError, ackErr.Error()))
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypePacket,
			eventAttributes...,
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	})
}

func EmitOnAckEvents(ctx sdk.Context, packetData types.FungibleTokenPacketDataV2, ack channeltypes.Acknowledgement) {
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypePacket,
			sdk.NewAttribute(sdk.AttributeKeySender, packetData.Sender),
			sdk.NewAttribute(types.AttributeKeyReceiver, packetData.Receiver),
			sdk.NewAttribute(types.AttributeKeyTokens, types.Tokens(packetData.Tokens).String()),
			sdk.NewAttribute(types.AttributeKeyMemo, packetData.Memo),
			sdk.NewAttribute(types.AttributeKeyAck, ack.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	})

	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)
	case *channeltypes.Acknowledgement_Error:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePacket,
				sdk.NewAttribute(types.AttributeKeyAckError, resp.Error),
			),
		)
	}
}

func EmitOnTimeoutEvents(ctx sdk.Context, packetData types.FungibleTokenPacketDataV2) {
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTimeout,
			sdk.NewAttribute(types.AttributeKeyReceiver, packetData.Sender),
			sdk.NewAttribute(types.AttributeKeyRefundTokens, types.Tokens(packetData.Tokens).String()),
			sdk.NewAttribute(types.AttributeKeyMemo, packetData.Memo),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	})
}

func emitTransferEvents(ctx sdk.Context, msg types.MsgTransfer) {
	coins := msg.GetCoins()

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTransfer,
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
			sdk.NewAttribute(types.AttributeKeyTokens, coins.String()),
			sdk.NewAttribute(types.AttributeKeyMemo, msg.Memo),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	})
}

func emitDenomTraceEvent(ctx sdk.Context, traceHash string, voucherDenom string) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDenomTrace,
			sdk.NewAttribute(types.AttributeKeyTraceHash, traceHash),
			sdk.NewAttribute(types.AttributeKeyDenom, voucherDenom),
		),
	)
}
