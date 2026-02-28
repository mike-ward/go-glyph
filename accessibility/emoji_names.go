package accessibility

// GetEmojiName returns a short name for common emoji codepoints.
// Returns empty string if the codepoint is unknown.
func GetEmojiName(ch rune) string {
	switch ch {
	// Smileys
	case 0x1F600:
		return "grinning face"
	case 0x1F601:
		return "beaming face"
	case 0x1F602:
		return "tears of joy"
	case 0x1F603:
		return "smiling face"
	case 0x1F604:
		return "grinning squinting face"
	case 0x1F605:
		return "grinning face with sweat"
	case 0x1F606:
		return "squinting face"
	case 0x1F607:
		return "smiling face with halo"
	case 0x1F609:
		return "winking face"
	case 0x1F60A:
		return "smiling face with smiling eyes"
	case 0x1F60B:
		return "face savoring food"
	case 0x1F60C:
		return "relieved face"
	case 0x1F60D:
		return "heart eyes"
	case 0x1F60E:
		return "sunglasses face"
	case 0x1F60F:
		return "smirking face"
	case 0x1F610:
		return "neutral face"
	case 0x1F611:
		return "expressionless face"
	case 0x1F612:
		return "unamused face"
	case 0x1F613:
		return "downcast face with sweat"
	case 0x1F614:
		return "pensive face"
	case 0x1F615:
		return "confused face"
	case 0x1F616:
		return "confounded face"
	case 0x1F617:
		return "kissing face"
	case 0x1F618:
		return "face blowing kiss"
	case 0x1F619:
		return "kissing face with smiling eyes"
	case 0x1F61A:
		return "kissing face with closed eyes"
	case 0x1F61B:
		return "face with tongue"
	case 0x1F61C:
		return "winking face with tongue"
	case 0x1F61D:
		return "squinting face with tongue"
	case 0x1F61E:
		return "disappointed face"
	case 0x1F61F:
		return "worried face"
	case 0x1F620:
		return "angry face"
	case 0x1F621:
		return "pouting face"
	case 0x1F622:
		return "crying face"
	case 0x1F623:
		return "persevering face"
	case 0x1F624:
		return "face with steam"
	case 0x1F625:
		return "sad but relieved face"
	case 0x1F626:
		return "frowning face with open mouth"
	case 0x1F627:
		return "anguished face"
	case 0x1F628:
		return "fearful face"
	case 0x1F629:
		return "weary face"
	case 0x1F62A:
		return "sleepy face"
	case 0x1F62B:
		return "tired face"
	case 0x1F62C:
		return "grimacing face"
	case 0x1F62D:
		return "loudly crying face"
	case 0x1F62E:
		return "face with open mouth"
	case 0x1F62F:
		return "hushed face"
	case 0x1F630:
		return "anxious face with sweat"
	case 0x1F631:
		return "face screaming"
	case 0x1F632:
		return "astonished face"
	case 0x1F633:
		return "flushed face"
	case 0x1F634:
		return "sleeping face"
	case 0x1F635:
		return "dizzy face"
	case 0x1F636:
		return "face without mouth"
	case 0x1F637:
		return "face with medical mask"
	// Gestures
	case 0x1F44D:
		return "thumbs up"
	case 0x1F44E:
		return "thumbs down"
	case 0x1F44F:
		return "clapping hands"
	case 0x1F64C:
		return "raising hands"
	case 0x1F64F:
		return "folded hands"
	case 0x270B:
		return "raised hand"
	case 0x270C:
		return "victory hand"
	case 0x1F44B:
		return "waving hand"
	case 0x1F44A:
		return "fist"
	case 0x1F91D:
		return "handshake"
	// Hearts
	case 0x2764:
		return "red heart"
	case 0x1F494:
		return "broken heart"
	case 0x1F495:
		return "two hearts"
	case 0x1F496:
		return "sparkling heart"
	case 0x1F497:
		return "growing heart"
	case 0x1F498:
		return "heart with arrow"
	case 0x1F499:
		return "blue heart"
	case 0x1F49A:
		return "green heart"
	case 0x1F49B:
		return "yellow heart"
	case 0x1F49C:
		return "purple heart"
	case 0x1F5A4:
		return "black heart"
	// Common symbols
	case 0x2705:
		return "check mark"
	case 0x274C:
		return "cross mark"
	case 0x2B50:
		return "star"
	case 0x1F525:
		return "fire"
	case 0x1F4A1:
		return "light bulb"
	case 0x1F389:
		return "party popper"
	case 0x1F680:
		return "rocket"
	case 0x1F4AF:
		return "hundred points"
	case 0x1F914:
		return "thinking face"
	case 0x1F923:
		return "rolling on floor laughing"
	}
	return ""
}
