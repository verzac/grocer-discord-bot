package handlers

func (m *MessageHandlerContext) OnReset() error {
	// if err := m.reply("Deleting all data for this server from my database... Please stand-by... :robot:"); err != nil {
	// 	return m.onError(err)
	// }
	if err := m.guildsService.ResetGuild(m.ctx, m.commandContext.GuildID); err != nil {
		return m.onError(err)
	}
	if err := m.reply(":wave: I've successfully deleted all of your data from my database! (p.s. you may need to set up commands such as /grohere or /developer again)"); err != nil {
		return m.onError(err)
	}
	return nil
}
