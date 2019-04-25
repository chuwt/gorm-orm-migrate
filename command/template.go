package command

var Template = "/* meta\n" +
	"time:%s\n" +
	"reversion:%s\n" +
	"down_revision:%s\n" +
	"*/\n" +
	"\n" +
	"-- upgrade\n" +
	"%s\n" +
	"-- end upgrade\n" +
	"\n" +
	"-- downgrade\n" +
	"%s\n" +
	"-- end downgrade"
