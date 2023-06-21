const editor_container = document.getElementById("editor-container");
const editor_content = document.getElementById("editor-content");
const mantis_editor = document.getElementById('editor');

let pos = 0, 
    length = 0,
    line_no = 0,
    no_lines = 1;

const bsKey     = 8,
      tabKey    = 9,
      enterKey  = 13,
      shiftKey  = 16,
      ctrlKey   = 17,
      altKey    = 18,
      capsKey   = 20,
      escKey    = 27,
      leftKey   = 37,
      upKey     = 38,
      rightKey  = 39,
      downKey   = 40,
      insertKey = 45, //fn return/enter
      delKey    = 46, //fn bs
      metaKey   = 91; //cmd


function new_editor_content_size() {
	editor_content.style.height = mantis_editor.offsetHeight + "px";
}
new ResizeObserver(() => new_editor_content_size()).observe(editor_content);

function scroll_editor_content() {
	editor_content.scrollTop = mantis_editor.scrollTop;
}

function redraw_editor() {
	editor_content.innerHTML = 
		mantis_editor.value.replaceAll(
			"&", "&amp;").replaceAll(
			"<", "&lt;").replaceAll(
			"-", "&#8288;-&#8288;") + '\n';

	microlight.reset([editor_content]);
}
//fixes newline issues when loading content with newlines
mantis_editor.innerHTML = 
	mantis_editor.innerText;/*.replaceAll(
		"\n\n", "<div><br></div>").replaceAll(
		"\n", "<div></div>") + '<div><br></div>';*/
//microlight.reset([editor_content]);
redraw_editor();

function insert_tab() {
	if(!window.getSelection)
		return;
	const sel = window.getSelection();

	if(!sel.rangeCount)
		return;
	const range = sel.getRangeAt(0);
	range.collapse(true);

	const span = document.createElement('span');
	span.appendChild(document.createTextNode('\t'));
	span.style.whiteSpace = 'pre';

	range.insertNode(span);
	// Move the caret immediately after the inserted span
	range.setStartAfter(span);
	range.collapse(true);
	sel.removeAllRanges();
	sel.addRange(range);
}

//const mantis_editor = document.getElementById("editor");
mantis_editor.innerHTML = mantis_editor.innerText;

function key_press(e) {
	//console.log("key_press");

	switch(e.keyCode) {
		case upKey:
			return;
		case downKey:
			return;
		case leftKey:
			return;
		case rightKey:
			return;
		case shiftKey:
			return;
		case ctrlKey:
			return;
		case metaKey:
			return;
		case altKey:
			return;
		case capsKey:
			return;
		case escKey:
			return;
	}


	if(!window.getSelection)
		return;
	const sel = window.getSelection();
	let currentNode = sel.focusNode.parentNode;
	let to_check = [mantis_editor];
	let pos = sel.focusOffset, childNo = 0;

	//console.log("focusNodeText: " + sel.focusNode.textContent);

	//does it matter if this is bfs or dfs?
	//presumably elements of lines can have children?
	for(let i=0; i<to_check.length; ++i) {
		for(let n=0; n<to_check[i].childNodes.length; ++n) {
			if(to_check[i].childNodes[n] == sel.focusNode) {
				i = to_check.length;
				break;
			}

			if(to_check[i].childNodes[n].nodeType == 3) {
				pos += to_check[i].childNodes[n].textContent.length;
				//console.log("editor kid: " + to_check[i].childNodes[n].textContent);
			}
			else {
				if(to_check[i].childNodes[n].childNodes.length)
					to_check.push(to_check[i].childNodes[n]);
				else
					++pos;
				//console.log("in here " + to_check[i].childNodes[n].childNodes.length);
			}
		}
	}
	
	e.preventDefault();

	let txt = mantis_editor.innerHTML;
	switch(e.keyCode) {
		case bsKey:
			if(!pos)
				return;
			--pos;
			if(pos < txt.length)
				mantis_editor.innerHTML = txt.slice(0, pos) + txt.slice(pos+1);
			else
				mantis_editor.innerHTML = txt.slice(0, pos);
			break;
		case tabKey:
			if(pos < txt.length)
				mantis_editor.innerHTML = txt.slice(0, pos) + '\t' + txt.slice(pos);
			else
				mantis_editor.innerHTML = txt.slice(0, pos) + '\t';
			++pos;
			break;
		case enterKey:
			if(pos < txt.length)
				mantis_editor.innerHTML = txt.slice(0, pos) + '\n' + txt.slice(pos);
			else
				mantis_editor.innerHTML = txt.slice(0, pos) + '\n';
			++pos;
			break;
		default:
			if(pos < txt.length)
				mantis_editor.innerHTML = txt.slice(0, pos) + e.key + txt.slice(pos);
			else
				mantis_editor.innerHTML = txt.slice(0, pos) + e.key;
			++pos;
	}

	redraw_editor();
	//microlight.reset([mantis_editor]);

	// Move the caret immediately after the inserted span
	if(!sel.rangeCount)
		return;
	const range = sel.getRangeAt(0);

	//does it matter if this is bfs or dfs?
	to_check = [mantis_editor];
	for(let i=0; i<to_check.length; ++i) {
		for(let n=0; n<to_check[i].childNodes.length; ++n) {
			if(to_check[i].childNodes[n].nodeType == 3) {
				if(pos <= to_check[i].childNodes[n].length) {
					//console.log("should have happened");
					if(e.keyCode == leftKey) {
						range.setStart(to_check[i].childNodes[n], pos);
						range.setEnd(  to_check[i].childNodes[n], pos);
					}
					else
						range.setStart(to_check[i].childNodes[n], pos);
					i = to_check.length;
					break;
				}
				else
					pos -= to_check[i].childNodes[n].textContent.length;
			}
			else {
				if(to_check[i].childNodes[n].childNodes.length)
					to_check.push(to_check[i].childNodes[n]);
				else if(pos)
					--pos;
				else {
					//console.log("should have happened3.0");
					if(e.keyCode == leftKey) {
						range.setStart(to_check[i].childNodes[n], 0);
						range.setEnd(  to_check[i].childNodes[n], 0);
					}
					else
						range.setStart(to_check[i].childNodes[n], 0);
					i = to_check.length;
					break;
				}
			}
		}
	}

    sel.removeAllRanges();
    sel.addRange(range);
}

	

/*document.getElementById("editor").addEventListener("keydown", (event) => {
    Promise.resolve().then(_ => {
        microlight.reset(document.getElementsByClassName("microlight"));
    });
});*/

//document.addEventListener('keydown', key_press, false);
//document.addEventListener('keyup'  , key_press, false);
