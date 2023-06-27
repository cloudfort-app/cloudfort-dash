const editor_container = document.getElementById("editor-container");
const editor_content = document.getElementById("editor-content");
const mantis_editor = document.getElementById('editor');

function new_editor_content_size() {
	editor_content.style.height = mantis_editor.offsetHeight + "px";
}
new ResizeObserver(() => new_editor_content_size()).observe(editor_content);

function scroll_editor_content() {
	editor_content.scrollTop = mantis_editor.scrollTop;
}

function scroll_editor() {
	mantis_editor.scrollTop = editor_content.scrollTop;
}

var highlight_editor = "true";
function redraw_editor() {
	if(highlight_editor == "true") {
		editor_content.innerHTML = hljs.highlightAuto(mantis_editor.value).value + "\n";
	}
	else {
		editor_content.innerHTML = 
			mantis_editor.value.replaceAll(
				"&", "&amp;").replaceAll(
				"<", "&lt;").replaceAll(
				"-", "&#8288;-&#8288;") + '\n';
	}
}
//fixes newline issues when loading content with newlines
mantis_editor.innerHTML = mantis_editor.innerText;
redraw_editor();

//const mantis_editor = document.getElementById("editor");
mantis_editor.innerHTML = mantis_editor.innerText;

//modified from https://stackoverflow.com/questions/6637341/use-tab-to-indent-in-textarea
mantis_editor.onkeydown = function(e) {
	// Tab key?
	if(e.keyCode === 9 || e.which==9) {
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