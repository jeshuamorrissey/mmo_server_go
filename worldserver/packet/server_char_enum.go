package packet

import (
	"bytes"
	"encoding/binary"

	"github.com/jeshuamorrissey/wow_server_go/worldserver/data/config"
	"github.com/jeshuamorrissey/wow_server_go/worldserver/data/static"
	"github.com/jeshuamorrissey/wow_server_go/worldserver/system"
	"github.com/sirupsen/logrus"
)

// ServerCharEnum is sent back in response to ClientPing.
type ServerCharEnum struct {
	Characters []*config.Character
}

// ToBytes writes out the packet to an array of bytes.
func (pkt *ServerCharEnum) ToBytes(state *system.State) ([]byte, error) {
	buffer := bytes.NewBufferString("")

	buffer.WriteByte(uint8(len(pkt.Characters))) // number of characters

	for _, char := range pkt.Characters {

		if !state.OM.Exists(char.GUID) {
			state.Log.Errorf("GameObject doesn't exist for character %v!", char.Name)
			continue
		}

		charObj := state.OM.GetPlayer(char.GUID)
		binary.Write(buffer, binary.LittleEndian, charObj.GUID().Low())
		binary.Write(buffer, binary.LittleEndian, charObj.GUID().High())
		buffer.WriteString(char.Name)
		buffer.WriteByte(0)
		buffer.WriteByte(uint8(charObj.Race.ID))
		buffer.WriteByte(uint8(charObj.Class.ID))
		buffer.WriteByte(uint8(charObj.Gender))
		buffer.WriteByte(uint8(charObj.SkinColor))
		buffer.WriteByte(uint8(charObj.Face))
		buffer.WriteByte(uint8(charObj.HairStyle))
		buffer.WriteByte(uint8(charObj.HairColor))
		buffer.WriteByte(uint8(charObj.Feature))
		buffer.WriteByte(uint8(charObj.Level))
		binary.Write(buffer, binary.LittleEndian, uint32(charObj.ZoneID))
		binary.Write(buffer, binary.LittleEndian, uint32(charObj.MapID))
		binary.Write(buffer, binary.LittleEndian, float32(charObj.GetLocation().X))
		binary.Write(buffer, binary.LittleEndian, float32(charObj.GetLocation().Y))
		binary.Write(buffer, binary.LittleEndian, float32(charObj.GetLocation().Z))

		// TODO(jeshua): implement the following fields with comments.
		binary.Write(buffer, binary.LittleEndian, uint32(0)) // GuildID
		binary.Write(buffer, binary.LittleEndian, char.Flags())

		if !char.HasLoggedIn {
			buffer.WriteByte(1)
		} else {
			buffer.WriteByte(0)
		}

		binary.Write(buffer, binary.LittleEndian, uint32(0)) // PetID
		binary.Write(buffer, binary.LittleEndian, uint32(0)) // PetLevel
		binary.Write(buffer, binary.LittleEndian, uint32(0)) // PetFamily

		for slot := static.EquipmentSlotHead; slot <= static.EquipmentSlotTabard; slot++ {
			if itemGUID, ok := charObj.Equipment[slot]; ok {
				if !state.OM.Exists(itemGUID) {
					state.Log.WithFields(logrus.Fields{
						"player":    charObj.GUID(),
						"slot":      slot,
						"item_guid": itemGUID,
					}).Errorf("Unknown equipped item")
					continue
				}

				item := state.OM.GetItem(itemGUID)
				binary.Write(buffer, binary.LittleEndian, uint32(item.GetTemplate().DisplayID))
				binary.Write(buffer, binary.LittleEndian, uint8(item.GetTemplate().InventoryType))
			} else {
				binary.Write(buffer, binary.LittleEndian, uint32(0))
				binary.Write(buffer, binary.LittleEndian, uint8(0))
			}
		}

		firstBag := charObj.FirstBag()
		if firstBag != nil {
			binary.Write(buffer, binary.LittleEndian, uint32(firstBag.GetTemplate().DisplayID))
			binary.Write(buffer, binary.LittleEndian, uint8(firstBag.GetTemplate().InventoryType))
		} else {
			binary.Write(buffer, binary.LittleEndian, uint32(0))
			binary.Write(buffer, binary.LittleEndian, uint8(0))

		}

	}

	return buffer.Bytes(), nil
}

// OpCode gets the opcode of the packet.
func (*ServerCharEnum) OpCode() static.OpCode {
	return static.OpCodeServerCharEnum
}
