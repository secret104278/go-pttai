// Copyright 2018 The go-pttai Authors
// This file is part of the go-pttai library.
//
// The go-pttai library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-pttai library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-pttai library. If not, see <http://www.gnu.org/licenses/>.

package e2e

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/me"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/stretchr/testify/assert"
	baloo "gopkg.in/h2non/baloo.v3"
)

func TestMultiDeviceUpdateSyncBoard(t *testing.T) {
	NNodes = 2
	isDebug := true

	var bodyString string
	var marshaled []byte
	var marshaledID []byte
	var marshaledID2 []byte
	var marshaledID3 []byte
	var marshaledStr string
	assert := assert.New(t)

	setupTest(t)
	defer teardownTest(t)

	t0 := baloo.New("http://127.0.0.1:9450")
	t1 := baloo.New("http://127.0.0.1:9451")

	// 1. get
	bodyString = `{"id": "testID", "method": "me_get", "params": []}`

	me0_1 := &me.BackendMyInfo{}
	testCore(t0, bodyString, me0_1, t, isDebug)
	assert.Equal(types.StatusAlive, me0_1.Status)

	me1_1 := &me.BackendMyInfo{}
	testCore(t1, bodyString, me1_1, t, isDebug)
	assert.Equal(types.StatusAlive, me1_1.Status)
	nodeID1_1 := me1_1.NodeID
	pubKey1_1, _ := nodeID1_1.Pubkey()
	nodeAddr1_1 := crypto.PubkeyToAddress(*pubKey1_1)

	// 3. getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": [""]}`

	me0_3 := &me.MyInfo{}
	testCore(t0, bodyString, me0_3, t, isDebug)
	assert.Equal(types.StatusAlive, me0_3.Status)
	assert.Equal(me0_1.ID, me0_3.ID)
	assert.Equal(1, len(me0_3.OwnerIDs))
	assert.Equal(me0_3.ID, me0_3.OwnerIDs[0])
	assert.Equal(true, me0_3.IsOwner(me0_3.ID))

	me1_3 := &me.MyInfo{}
	testCore(t1, bodyString, me1_3, t, isDebug)
	assert.Equal(types.StatusAlive, me1_3.Status)
	assert.Equal(me1_1.ID, me1_3.ID)
	assert.Equal(1, len(me1_3.OwnerIDs))
	assert.Equal(me1_3.ID, me1_3.OwnerIDs[0])
	assert.Equal(true, me1_3.IsOwner(me1_3.ID))

	// 4. show-my-key
	bodyString = `{"id": "testID", "method": "me_showMyKey", "params": []}`

	var myKey0_4 string

	testCore(t0, bodyString, &myKey0_4, t, isDebug)
	if isDebug {
		t.Logf("myKey0_4: %v\n", myKey0_4)
	}

	// 5. show-me-url
	bodyString = `{"id": "testID", "method": "me_showMeURL", "params": []}`

	dataShowMeURL1_5 := &pkgservice.BackendJoinURL{}
	testCore(t1, bodyString, dataShowMeURL1_5, t, isDebug)
	meURL1_5 := dataShowMeURL1_5.URL

	// 6. me_GetMyNodes
	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes0_6 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMyNodes0_6, t, isDebug)
	assert.Equal(1, len(dataGetMyNodes0_6.Result))

	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes1_6 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMyNodes1_6, t, isDebug)
	assert.Equal(1, len(dataGetMyNodes1_6.Result))

	// 6.1 getJoinKeys
	bodyString = `{"id": "testID", "method": "me_getJoinKeyInfos", "params": [""]}`
	dataGetJoinKeys0_6_1 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetJoinKeys0_6_1, t, isDebug)
	assert.Equal(1, len(dataGetJoinKeys0_6_1.Result))

	// 6.2 create-board
	title := []byte("板名1_1")
	marshaledStr = base64.StdEncoding.EncodeToString(title)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createBoard", "params": ["%v", true]}`, marshaledStr)

	dataCreateBoard1_6_2 := &content.BackendCreateBoard{}

	testCore(t1, bodyString, dataCreateBoard1_6_2, t, isDebug)
	assert.Equal(pkgservice.EntityTypePrivate, dataCreateBoard1_6_2.BoardType)
	assert.Equal(title, dataCreateBoard1_6_2.Title)
	assert.Equal(types.StatusAlive, dataCreateBoard1_6_2.Status)
	assert.Equal(me1_3.ID, dataCreateBoard1_6_2.CreatorID)
	assert.Equal(me1_3.ID, dataCreateBoard1_6_2.UpdaterID)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 6.3 get board
	marshaledID, _ = dataCreateBoard1_6_2.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getRawBoard", "params": ["%v"]}`, string(marshaledID))

	board1_6_3 := &content.Board{}

	testCore(t1, bodyString, board1_6_3, t, isDebug)
	assert.Equal(board1_6_3.ID[common.AddressLength:], me1_3.ID[:common.AddressLength])
	assert.Equal(board1_6_3.CreatorID, me1_3.ID)
	assert.Equal(types.StatusAlive, board1_6_3.Status)
	assert.Equal(pkgservice.EntityTypePrivate, board1_6_3.EntityType)

	// 6.4 create-article
	article, _ := json.Marshal([]string{
		base64.StdEncoding.EncodeToString([]byte("文章內容1_1")),
		base64.StdEncoding.EncodeToString([]byte("文章內容1_2")),
	})

	marshaled, _ = dataCreateBoard1_6_2.ID.MarshalText()

	title1_6_4 := []byte("文章標題1_1")
	marshaledStr = base64.StdEncoding.EncodeToString(title1_6_4)

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_createArticle", "params": ["%v", "%v", %v, []]}`, string(marshaled), marshaledStr, string(article))
	dataCreateArticle1_6_4 := &content.BackendCreateArticle{}
	testCore(t1, bodyString, dataCreateArticle1_6_4, t, isDebug)
	assert.Equal(dataCreateBoard1_6_2.ID, dataCreateArticle1_6_4.BoardID)
	assert.Equal(2, dataCreateArticle1_6_4.NBlock)

	// 6.5 content-get-article-list
	marshaled, _ = dataCreateBoard1_6_2.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaled))
	dataGetArticleList1_6_5 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_6_5, t, isDebug)
	assert.Equal(1, len(dataGetArticleList1_6_5.Result))
	article1_6_5 := dataGetArticleList1_6_5.Result[0]
	assert.Equal(types.StatusAlive, article1_6_5.Status)

	// 6.6 get-article-block
	marshaled, _ = dataCreateBoard1_6_2.ID.MarshalText()
	marshaledID2, _ = article1_6_5.ID.MarshalText()
	marshaledID3, _ = article1_6_5.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaled), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList1_6_6 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_6_6, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList1_6_6.Result))

	article1_0 := [][]byte{
		[]byte("文章內容1_1"),
	}

	article1_1 := [][]byte{
		[]byte("文章內容1_2"),
	}

	assert.Equal(article1_0, dataGetArticleBlockList1_6_6.Result[0].Buf)
	assert.Equal(article1_1, dataGetArticleBlockList1_6_6.Result[1].Buf)

	// 49. update-article
	// article48, _ := json.Marshal([]string{
	// 	base64.StdEncoding.EncodeToString([]byte("測試61")),
	// 	base64.StdEncoding.EncodeToString([]byte("測試62")),
	// 	base64.StdEncoding.EncodeToString([]byte("測試63")),
	// 	base64.StdEncoding.EncodeToString([]byte("測試64")),
	// 	base64.StdEncoding.EncodeToString([]byte("測試65")),
	// 	base64.StdEncoding.EncodeToString([]byte("測試66")),
	// 	base64.StdEncoding.EncodeToString([]byte("測試67")),
	// 	base64.StdEncoding.EncodeToString([]byte("測試68")),
	// 	base64.StdEncoding.EncodeToString([]byte("測試69")),
	// })

	// marshaledID, _ = dataCreateBoard1_6_2.ID.MarshalText()
	// marshaledID2, _ = article1_6_5.ID.MarshalText()

	// bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_updateArticle", "params": ["%v", "%v", %v, []]}`, string(marshaledID), string(marshaledID2), string(article48))
	// dataUpdateArticle1_48 := &content.BackendUpdateArticle{}
	// testCore(t1, bodyString, dataUpdateArticle1_48, t, isDebug)
	// assert.Equal(dataCreateBoard1_6_2.ID, dataUpdateArticle1_48.BoardID)
	// assert.Equal(article1_6_5.ID, dataUpdateArticle1_48.ArticleID)
	// assert.Equal(2, dataUpdateArticle1_48.NBlock)

	// wait 10 seconds
	time.Sleep(10 * time.Second)

	// 49. content-get-article-list
	marshaledID, _ = dataCreateBoard1_6_2.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList1_49 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_49, t, isDebug)
	assert.Equal(1, len(dataGetArticleList1_49.Result))
	article1_49_0 := dataGetArticleList1_49.Result[0]
	assert.Equal(types.StatusAlive, article1_49_0.Status)
	//assert.Equal(dataUpdateArticle1_48.ContentBlockID, article1_49_0.ContentBlockID)

	// 50. get-article-block
	marshaledID, _ = dataCreateBoard1_6_2.ID.MarshalText()
	marshaledID2, _ = article1_6_5.ID.MarshalText()
	marshaledID3, _ = article1_6_5.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	dataGetArticleBlockList1_50 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_50, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList1_50.Result))

	// article50_0 := [][]byte{
	// 	[]byte("測試61"),
	// }

	// article50_1 := [][]byte{
	// 	[]byte("測試62"),
	// 	[]byte("測試63"),
	// 	[]byte("測試64"),
	// 	[]byte("測試65"),
	// 	[]byte("測試66"),
	// 	[]byte("測試67"),
	// 	[]byte("測試68"),
	// 	[]byte("測試69"),
	// }

	// assert.Equal(article50_0, dataGetArticleBlockList1_50.Result[0].Buf)
	// assert.Equal(article50_1, dataGetArticleBlockList1_50.Result[1].Buf)

	// 7. join-me t0 signed-in as t1
	log.Debug("7. join-me")

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_joinMe", "params": ["%v", "%v", false]}`, meURL1_5, myKey0_4)

	dataJoinMe0_7 := &pkgservice.BackendJoinRequest{}
	testCore(t0, bodyString, dataJoinMe0_7, t, true)

	assert.Equal(me1_3.ID, dataJoinMe0_7.CreatorID)
	assert.Equal(me1_1.NodeID, dataJoinMe0_7.NodeID)

	// wait 10
	t.Logf("wait 10 seconds for hand-shaking")
	time.Sleep(TimeSleepRestart)

	// wait 10 secs
	time.Sleep(10 * time.Second)

	// 8. me_GetMyNodes
	bodyString = `{"id": "testID", "method": "me_getMyNodes", "params": []}`
	dataGetMyNodes0_8 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetMyNodes0_8, t, isDebug)
	assert.Equal(2, len(dataGetMyNodes0_8.Result))
	myNode0_8_0 := dataGetMyNodes0_8.Result[0]
	myNode0_8_1 := dataGetMyNodes0_8.Result[1]

	assert.Equal(types.StatusAlive, myNode0_8_0.Status)
	assert.Equal(types.StatusAlive, myNode0_8_1.Status)

	dataGetMyNodes1_8 := &struct {
		Result []*me.MyNode `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetMyNodes1_8, t, isDebug)
	assert.Equal(2, len(dataGetMyNodes1_8.Result))
	myNode1_8_0 := dataGetMyNodes1_8.Result[0]
	myNode1_8_1 := dataGetMyNodes1_8.Result[1]

	assert.Equal(types.StatusAlive, myNode1_8_0.Status)
	assert.Equal(types.StatusAlive, myNode1_8_1.Status)

	// 8.1. getRawMe
	bodyString = `{"id": "testID", "method": "me_getRawMe", "params": [""]}`

	me0_8_1 := &me.MyInfo{}
	testCore(t0, bodyString, me0_8_1, t, isDebug)
	assert.Equal(types.StatusAlive, me0_8_1.Status)
	assert.Equal(1, len(me0_8_1.OwnerIDs))
	assert.Equal(me1_3.ID, me0_8_1.OwnerIDs[0])
	assert.Equal(true, me0_8_1.IsOwner(me1_3.ID))

	me1_8_1 := &me.MyInfo{}
	testCore(t1, bodyString, me1_8_1, t, isDebug)
	assert.Equal(types.StatusAlive, me1_8_1.Status)
	assert.Equal(me1_3.ID, me1_8_1.ID)
	assert.Equal(1, len(me1_8_1.OwnerIDs))
	assert.Equal(me1_3.ID, me1_8_1.OwnerIDs[0])
	assert.Equal(true, me1_8_1.IsOwner(me1_3.ID))

	// 9. MasterOplog
	bodyString = `{"id": "testID", "method": "me_getMyMasterOplogList", "params": ["", "", 0, 2]}`

	dataMasterOplogs0_9 := &struct {
		Result []*me.MasterOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMasterOplogs0_9, t, isDebug)
	assert.Equal(3, len(dataMasterOplogs0_9.Result))
	masterOplog0_9 := dataMasterOplogs0_9.Result[0]
	assert.Equal(me1_3.ID[:common.AddressLength], masterOplog0_9.CreatorID[common.AddressLength:])
	assert.Equal(me1_3.ID, masterOplog0_9.ObjID)
	assert.Equal(me.MasterOpTypeAddMaster, masterOplog0_9.Op)
	assert.Equal(nilPttID, masterOplog0_9.PreLogID)
	assert.Equal(types.Bool(true), masterOplog0_9.IsSync)
	assert.Equal(masterOplog0_9.ID, masterOplog0_9.MasterLogID)

	dataMasterOplogs1_9 := &struct {
		Result []*me.MasterOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMasterOplogs1_9, t, isDebug)
	assert.Equal(3, len(dataMasterOplogs1_9.Result))
	masterOplog1_9 := dataMasterOplogs1_9.Result[0]
	assert.Equal(me1_3.ID[:common.AddressLength], masterOplog1_9.CreatorID[common.AddressLength:])
	assert.Equal(me1_3.ID, masterOplog1_9.ObjID)
	assert.Equal(me.MasterOpTypeAddMaster, masterOplog1_9.Op)
	assert.Equal(nilPttID, masterOplog1_9.PreLogID)
	assert.Equal(types.Bool(true), masterOplog1_9.IsSync)
	assert.Equal(masterOplog1_9.ID, masterOplog1_9.MasterLogID)

	//masterOplog1_9_2 := dataMasterOplogs1_9.Result[2]

	for i, oplog := range dataMasterOplogs0_9.Result {
		oplog1 := dataMasterOplogs1_9.Result[i]
		oplog.CreateTS = oplog1.CreateTS
		oplog.CreatorID = oplog1.CreatorID
		oplog.CreatorHash = oplog1.CreatorHash
		oplog.Salt = oplog1.Salt
		oplog.Sig = oplog1.Sig
		oplog.Pubkey = oplog1.Pubkey
		oplog.KeyExtra = oplog1.KeyExtra
		oplog.UpdateTS = oplog1.UpdateTS
		oplog.Hash = oplog1.Hash
		oplog.IsNewer = oplog1.IsNewer
		oplog.Extra = oplog1.Extra
	}
	assert.Equal(dataMasterOplogs0_9, dataMasterOplogs1_9)

	// 9.1. getRawMe
	marshaled, _ = me0_3.ID.MarshalText()
	bodyString = fmt.Sprintf(`{"id": "testID", "method": "me_getRawMe", "params": ["%v"]}`, string(marshaled))

	me0_9_1 := &me.MyInfo{}
	testCore(t0, bodyString, me0_9_1, t, isDebug)
	assert.Equal(types.StatusMigrated, me0_9_1.Status)
	assert.Equal(2, len(me0_9_1.OwnerIDs))
	assert.Equal(true, me0_9_1.IsOwner(me1_3.ID))
	assert.Equal(true, me0_9_1.IsOwner(me0_3.ID))

	// 9.2. MeOplog
	bodyString = `{"id": "testID", "method": "me_getMeOplogList", "params": ["", 0, 2]}`

	dataMeOplogs0_9_2 := &struct {
		Result []*me.MeOplog `json:"result"`
	}{}
	testListCore(t0, bodyString, dataMeOplogs0_9_2, t, isDebug)
	assert.Equal(2, len(dataMeOplogs0_9_2.Result))
	meOplog0_9_2_1 := dataMeOplogs0_9_2.Result[0]
	meOplog0_9_2_2 := dataMeOplogs0_9_2.Result[1]

	assert.Equal(me1_3.ID, meOplog0_9_2_1.CreatorID)
	assert.Equal(me1_3.ID, meOplog0_9_2_2.CreatorID)
	assert.Equal(me1_3.ID, meOplog0_9_2_1.ObjID)
	assert.Equal(dataCreateBoard1_6_2.ID, meOplog0_9_2_2.ObjID)
	assert.Equal(me.MeOpTypeCreateMe, meOplog0_9_2_1.Op)
	assert.Equal(me.MeOpTypeCreateBoard, meOplog0_9_2_2.Op)
	assert.Equal(nilPttID, meOplog0_9_2_1.PreLogID)
	assert.Equal(nilPttID, meOplog0_9_2_2.PreLogID)
	assert.Equal(types.Bool(true), meOplog0_9_2_1.IsSync)
	//assert.Equal(types.Bool(true), meOplog0_9_2_2.IsSync)
	assert.Equal(masterOplog1_9.ID, meOplog0_9_2_1.MasterLogID)
	assert.Equal(masterOplog1_9.ID, meOplog0_9_2_2.MasterLogID)
	assert.Equal(me1_3.LogID, meOplog0_9_2_1.ID)
	//assert.Equal(me1_3.LogID, meOplog0_9_2_2.ID)
	masterSign0_9_2 := meOplog0_9_2_1.MasterSigns[0]
	//assert.Equal(nodeAddr0_1[:], meOplog0_9_2_1.ID[:common.AddressLength])
	assert.Equal(me1_3.ID[:common.AddressLength], masterSign0_9_2.ID[common.AddressLength:])
	assert.Equal(me0_8_1.LogID, meOplog0_9_2_1.ID)

	dataMeOplogs1_9_2 := &struct {
		Result []*me.MeOplog `json:"result"`
	}{}
	testListCore(t1, bodyString, dataMeOplogs1_9_2, t, isDebug)
	assert.Equal(2, len(dataMeOplogs1_9_2.Result))
	meOplog1_9_2_1 := dataMeOplogs1_9_2.Result[0]
	meOplog1_9_2_2 := dataMeOplogs1_9_2.Result[1]
	assert.Equal(me1_3.ID, meOplog1_9_2_1.CreatorID)
	assert.Equal(me1_3.ID, meOplog1_9_2_2.CreatorID)
	assert.Equal(me1_3.ID, meOplog1_9_2_1.ObjID)
	assert.Equal(dataCreateBoard1_6_2.ID, meOplog1_9_2_2.ObjID)
	assert.Equal(me.MeOpTypeCreateMe, meOplog1_9_2_1.Op)
	assert.Equal(me.MeOpTypeCreateBoard, meOplog1_9_2_2.Op)
	assert.Equal(nilPttID, meOplog1_9_2_1.PreLogID)
	assert.Equal(nilPttID, meOplog1_9_2_2.PreLogID)
	assert.Equal(types.Bool(true), meOplog1_9_2_1.IsSync)
	assert.Equal(types.Bool(true), meOplog1_9_2_2.IsSync)
	assert.Equal(masterOplog1_9.ID, meOplog1_9_2_1.MasterLogID)
	assert.Equal(masterOplog1_9.ID, meOplog1_9_2_2.MasterLogID)
	assert.Equal(me1_3.LogID, meOplog1_9_2_1.ID)
	//assert.Equal(me1_3.LogID, meOplog1_9_2_2.ID)
	masterSign1_9_2 := meOplog1_9_2_1.MasterSigns[0]
	assert.Equal(nodeAddr1_1[:], masterSign1_9_2.ID[:common.AddressLength])
	assert.Equal(me1_3.ID[:common.AddressLength], masterSign1_9_2.ID[common.AddressLength:])
	assert.Equal(me1_8_1.LogID, meOplog1_9_2_1.ID)

	// 12. content-get-article-list
	marshaledID, _ = dataCreateBoard1_6_2.ID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleList", "params": ["%v", "", 0, 2]}`, string(marshaledID))
	dataGetArticleList0_12 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t0, bodyString, dataGetArticleList0_12, t, isDebug)
	assert.Equal(1, len(dataGetArticleList0_12.Result))
	article0_12_0 := dataGetArticleList0_12.Result[0]
	assert.Equal(types.StatusAlive, article0_12_0.Status)
	assert.Equal(article1_6_5.ContentBlockID, article0_12_0.ContentBlockID)

	dataGetArticleList1_12 := &struct {
		Result []*content.BackendGetArticle `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleList1_12, t, isDebug)
	assert.Equal(1, len(dataGetArticleList1_12.Result))
	article1_12_0 := dataGetArticleList1_12.Result[0]
	assert.Equal(types.StatusAlive, article1_12_0.Status)
	assert.Equal(article1_49_0.ID, article1_12_0.ID)
	assert.Equal(article1_49_0.ContentBlockID, article1_12_0.ContentBlockID)

	// 13. get-article-block
	marshaledID, _ = dataCreateBoard1_6_2.ID.MarshalText()
	marshaledID2, _ = article1_6_5.ID.MarshalText()
	marshaledID3, _ = article1_6_5.ContentBlockID.MarshalText()

	bodyString = fmt.Sprintf(`{"id": "testID", "method": "content_getArticleBlockList", "params": ["%v", "%v", "%v", 0, 0, 10, 2]}`, string(marshaledID), string(marshaledID2), string(marshaledID3))

	// article13_0 := [][]byte{
	// 	[]byte("測試61"),
	// }

	// article13_1 := [][]byte{
	// 	[]byte("測試62"),
	// 	[]byte("測試63"),
	// 	[]byte("測試64"),
	// 	[]byte("測試65"),
	// 	[]byte("測試66"),
	// 	[]byte("測試67"),
	// 	[]byte("測試68"),
	// 	[]byte("測試69"),
	// }

	// dataGetArticleBlockList0_13 := &struct {
	// 	Result []*content.ArticleBlock `json:"result"`
	// }{}
	// testListCore(t0, bodyString, dataGetArticleBlockList0_13, t, isDebug)
	// assert.Equal(2, len(dataGetArticleBlockList0_13.Result))

	// assert.Equal(article13_0, dataGetArticleBlockList0_13.Result[0].Buf)
	// assert.Equal(article13_1, dataGetArticleBlockList0_13.Result[1].Buf)

	dataGetArticleBlockList1_13 := &struct {
		Result []*content.ArticleBlock `json:"result"`
	}{}
	testListCore(t1, bodyString, dataGetArticleBlockList1_13, t, isDebug)
	assert.Equal(2, len(dataGetArticleBlockList1_13.Result))

	// assert.Equal(article13_0, dataGetArticleBlockList1_13.Result[0].Buf)
	// assert.Equal(article13_1, dataGetArticleBlockList1_13.Result[1].Buf)
}
