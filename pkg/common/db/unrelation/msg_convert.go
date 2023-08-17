package unrelation

import (
	"context"
	"fmt"

	"github.com/OpenIMSDK/tools/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	table "github.com/OpenIMSDK/Open-IM-Server/pkg/common/db/table/unrelation"
)

func (m *MsgMongoDriver) ConvertMsgsDocLen(ctx context.Context, conversationIDs []string) {
	for _, conversationID := range conversationIDs {
		regex := primitive.Regex{Pattern: fmt.Sprintf("^%s:", conversationID)}
		cursor, err := m.MsgCollection.Find(ctx, bson.M{"doc_id": regex})
		if err != nil {
			log.ZError(ctx, "convertAll find msg doc failed", err, "conversationID", conversationID)
			continue
		}
		var msgDocs []table.MsgDocModel
		err = cursor.All(ctx, &msgDocs)
		if err != nil {
			log.ZError(ctx, "convertAll cursor all failed", err, "conversationID", conversationID)
			continue
		}
		if len(msgDocs) < 1 {
			continue
		}
		log.ZInfo(ctx, "msg doc convert", "conversationID", conversationID, "len(msgDocs)", len(msgDocs))
		if len(msgDocs[0].Msg) == int(m.model.GetSingleGocMsgNum5000()) {
			if _, err := m.MsgCollection.DeleteMany(ctx, bson.M{"doc_id": regex}); err != nil {
				log.ZError(ctx, "convertAll delete many failed", err, "conversationID", conversationID)
				continue
			}
			var newMsgDocs []interface{}
			for _, msgDoc := range msgDocs {
				if int64(len(msgDoc.Msg)) == m.model.GetSingleGocMsgNum() {
					continue
				}
				var index int64
				for index < int64(len(msgDoc.Msg)) {
					msg := msgDoc.Msg[index]
					if msg != nil && msg.Msg != nil {
						msgDocModel := table.MsgDocModel{DocID: m.model.GetDocID(conversationID, msg.Msg.Seq)}
						end := index + m.model.GetSingleGocMsgNum()
						if int(end) >= len(msgDoc.Msg) {
							msgDocModel.Msg = msgDoc.Msg[index:]
						} else {
							msgDocModel.Msg = msgDoc.Msg[index:end]
						}
						newMsgDocs = append(newMsgDocs, msgDocModel)
						index = end
					} else {
						break
					}
				}
			}
			_, err = m.MsgCollection.InsertMany(ctx, newMsgDocs)
			if err != nil {
				log.ZError(ctx, "convertAll insert many failed", err, "conversationID", conversationID, "len(newMsgDocs)", len(newMsgDocs))
			} else {
				log.ZInfo(ctx, "msg doc convert", "conversationID", conversationID, "len(newMsgDocs)", len(newMsgDocs))
			}
		}
	}
}