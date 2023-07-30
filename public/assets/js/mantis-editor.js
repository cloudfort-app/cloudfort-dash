const editor_container = document.getElementById("editor-container");
const editor_content = document.getElementById("editor-content");
const mantis_editor = document.getElementById('editor');

function new_editor_content_size() {
	editor_content.style.height = mantis_editor.offsetHeight + "px";
}
new ResizeObserver(() => new_editor_content_size()).observe(editor_content);

function scroll_editor_content() {
	editor_content.scrollTop = mantis_editor.scrollTop;
	editor_content.scrollLeft = mantis_editor.scrollLeft;
}

function scroll_editor() {
	mantis_editor.scrollTop = editor_content.scrollTop;
	mantis_editor.scrollLeft = editor_content.scrollLeft;
}

function getLineHeight(el) {
    var temp = document.createElement(el.nodeName), ret;
    temp.setAttribute("style", "margin:0; padding:0; "
        + "font-family:" + (el.style.fontFamily || "inherit") + "; "
        + "font-size:" + (el.style.fontSize || "inherit"));
    temp.innerHTML = "A";

    el.parentNode.appendChild(temp);
    ret = temp.clientHeight;
    temp.parentNode.removeChild(temp);

    return ret;
}

function set_line_no() {
	//mantis_editor.title = "line " + mantis_editor.value.substr(0, mantis_editor.selectionStart).split("\n").length;
	document.getElementById("select-open-file").title = "line " + mantis_editor.value.substr(0, mantis_editor.selectionStart).split("\n").length;
}

function setSelectionRange(input, selectionStart, selectionEnd) {
    if (input.setSelectionRange) {
        input.focus();
        input.setSelectionRange(selectionStart, selectionEnd);
    }
    else if (input.createTextRange) {
        var range = input.createTextRange();
        range.collapse(true);
        range.moveEnd('character', selectionEnd);
        range.moveStart('character', selectionStart);
        range.select();
    }
}

function setCaretToPos (input, pos) {
	setSelectionRange(input, pos, pos);
}

function editor_duplicate_line() {
	//const selObj = window.getSelection();
	//const selRange = selObj.getRangeAt(0);

	if(mantis_editor.selectionStart == mantis_editor.selectionEnd) {
		const scrollTop = mantis_editor.scrollTop;
		var pos = mantis_editor.selectionStart;

		mantis_editor.value = mantis_editor.value.substring(0, mantis_editor.value.indexOf("\n", pos)) + mantis_editor.value.substring(mantis_editor.value.lastIndexOf("\n", pos-1), mantis_editor.value.indexOf("\n", pos)) + mantis_editor.value.substring(mantis_editor.value.indexOf("\n", pos), mantis_editor.value.length);

		pos += mantis_editor.value.indexOf("\n", pos) - mantis_editor.value.lastIndexOf("\n", pos-1);
		setCaretToPos(mantis_editor, pos);
		redraw_editor();
		mantis_editor.scrollTop = scrollTop;
	}
	else {
		const scrollTop = mantis_editor.scrollTop;
		const selStart = mantis_editor.selectionStart;
		const selEnd = mantis_editor.selectionEnd;
		const selText = window.getSelection().toString();
		mantis_editor.value = mantis_editor.value.substring(0, selEnd) + selText + mantis_editor.value.substring(selEnd, mantis_editor.value.length);
		setSelectionRange(mantis_editor, selStart, selEnd);
		redraw_editor();
		mantis_editor.scrollTop = scrollTop;
	}

	mark_edited();
}

function jump_to_line() {
	var line = document.getElementById("input-line-no").value;

	var pos = 0;
	var lines = mantis_editor.value.split("\n");
	for(var i=0; i+1<line; ++i)
		pos += lines[i].length + 1;
	setCaretToPos(mantis_editor, pos);

	var lineHeight = 14.4;
	var jump = (line - 1) * lineHeight;
	mantis_editor.scrollTop = jump;

	set_line_no();

	document.getElementById("card-input-line-no").style.display = "none";
}

function find_and_replace() {
	var find = document.getElementById("input-editor-find").value;
	var replace = document.getElementById("input-editor-replace").value;
	mantis_editor.value = mantis_editor.value.replaceAll(find, replace);
	redraw_editor();
	mark_edited();
}

function regex_replace() {
	var pattern = document.getElementById("input-editor-regex-pattern").value.replaceAll("\\", "\\\\");
	var flags = document.getElementById("input-editor-regex-flags").value.replaceAll("\\", "\\\\");
	var replace = document.getElementById("input-editor-regex-replace").value;
	var re = new RegExp(pattern, flags);
	mantis_editor.value = mantis_editor.value.replaceAll(re, replace);
	redraw_editor();
	mark_edited();
}

hljs.highlightAll();
hljs.initLineNumbersOnLoad();

function one_newline(str) {
	var count = 0;
	for(var i=0; count < 2 && i<str.length; i++) {
		if(str[i] == '\n')
			++count;
	}

	return (count == 1);
}

var highlight_editor = "true";
function redraw_editor() {
	set_line_no();

	if(highlight_editor == "true") {
		editor_content.innerHTML = hljs.highlightAuto(mantis_editor.value).value + "\n";
		if(one_newline(editor_content.innerHTML))
			editor_content.innerHTML = editor_content.innerHTML + "\n";
		editor_content.innerHTML = hljs.lineNumbersValue(editor_content.innerHTML);
	}
	else {
		editor_content.innerHTML = 
			mantis_editor.value.replaceAll(
				"&", "&amp;").replaceAll(
				"<", "&lt;").replaceAll(
				">", "&gt;").replaceAll(
				"-", "&#8288;-&#8288;") + '\n';
		if(one_newline(editor_content.innerHTML))
			editor_content.innerHTML = editor_content.innerHTML + "\n";

		editor_content.innerHTML = hljs.lineNumbersValue(editor_content.innerHTML);
	}
}
//fixes newline issues when loading content with newlines
mantis_editor.innerHTML = mantis_editor.innerText;
redraw_editor();

//const mantis_editor = document.getElementById("editor");
mantis_editor.innerHTML = mantis_editor.innerText;

//modified from https://stackoverflow.com/questions/6637341/use-tab-to-indent-in-textarea
mantis_editor.onkeydown = function(e) {
	var name = e.key;
	var code = e.code;

	if(e.altKey) {
		if(name == 'w') {
			e.preventDefault();
			close_file();
		}
		else if(name == '-' || name == '_') {
			e.preventDefault();

			for(var i=1; i<file_select.length; ++i) {
				if(file_select.options[i].value == open_path) {
					if(i>1)
						file_select.value = file_select.options[i-1].value;
					else
						file_select.value = file_select.options[file_select.length-1].value;

					select_open_file_event();
					break;
				}
			}
		}
		else if(name == '=' || name == '+') {
			e.preventDefault();

			for(var i=1; i<file_select.length; ++i) {
				if(file_select.options[i].value == open_path) {
					if(i + 1 < file_select.length) {
						file_select.value = file_select.options[i+1].value;
					}
					else {
						file_select.value = file_select.options[1].value;
					}

					select_open_file_event();
					break;
				}
			}
		}
	}
	if(e.ctrlKey) {
		if(name == 's') {
			e.preventDefault();
			save_file();
		}
		else if(name == "S") {
			e.preventDefault();
			show_save_as();
		}
		else if(name == "D") {
			e.preventDefault();
			editor_duplicate_line();
		}
		else if(name == 'F') {
			e.preventDefault();
			show_find_and_replace();
		}
		else if(name == 'X') {
			e.preventDefault();
			show_regex_replace();
		}
	}
	else if(e.keyCode === 9 || e.which==9) { // Tab key
		e.preventDefault();

		// selection?
		if(this.selectionStart == this.selectionEnd){
			// These single character operations are undoable
			if(!e.shiftKey)
				document.execCommand('insertText', false, "\t");
			else {
				var text = this.value;
				if (this.selectionStart > 0 && text[this.selectionStart-1]=='\t')
					document.execCommand('delete');
			}
		}
		else {
			// Block indent/unindent trashes undo stack.
			// Select whole lines
			var selStart = this.selectionStart;
			var selEnd = this.selectionEnd;
			var text = this.value;
			while (selStart > 0 && text[selStart-1] != '\n')
				selStart--;
			while (selEnd > 0 && text[selEnd-1]!='\n' && selEnd < text.length)
				selEnd++;

			// Get selected text
			var lines = text.substr(selStart, selEnd - selStart).split('\n');

			// Insert tabs
			for(var i=0; i<lines.length; i++) {
				// Don't indent last line if cursor at start of line
				if (i==lines.length-1 && lines[i].length==0)
					continue;

				// Tab or Shift+Tab?
				if(e.shiftKey) {
					if (lines[i].startsWith('\t'))
						lines[i] = lines[i].substr(1);
					else if (lines[i].startsWith("    "))
						lines[i] = lines[i].substr(4);
				}
				else
					lines[i] = "\t" + lines[i];
			}
			lines = lines.join('\n');

			// Update the text area
			this.value = text.substr(0, selStart) + lines + text.substr(selEnd);
			this.selectionStart = selStart;
			this.selectionEnd = selStart + lines.length; 
		}

		redraw_editor();
	}
}
