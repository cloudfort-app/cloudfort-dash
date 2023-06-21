/**
 * @fileoverview mantis-highlight - syntax highlightning library
 * @version 0.0.1
 *
 * @license MIT, see http://github.com/mantis-framework
 * @copyright 2022-present nicholas ham
 *
 * Forked from microlight
 */

function is_open_brace(c) {
	return /^({|\[|\()/['test'](c);
}

function is_close_brace(c) {
	return /^(}|]|\))/['test'](c);
}

function is_brace(c) {
	return /^({|}|\[|]|\(|\))/['test'](c);
    //return (is_open_brace(c) || is_close_brace(c));
    //return(a(c)||b(c));
    //check to see when each possibility will be smaller minified
}

function is_inline_whitespace(c) {
	return /^(' '|\t)/['test'](c);
}

function is_whitespace(c) {
	return /^(\s)/['test'](c);
}

function is_alphabetical(c) {
	return /^(\w)/['test'](c);
}

function is_numerical(c) {
	return /^(\d)/['test'](c);
}

function is_alphanumeric(c) {
	return /^(\d|\w)/['test'](c);
}

function is_c_var_name_char(c, is_start) {
	return is_start 
		? /^(\d|\w|_)/['test'](c) 
		: /^(\w|_)/['test'](c);
	/*return (
		c == '_' || (
			is_start 
				? is_alphanumeric(c) 
				: is_alphabetical(c)
		)
	);*/
}

function is_var_name_char(c, is_start) {
	return is_start 
		? /^(\d|\w|_|-)/['test'](c) 
		: /^(\w|_|-)/['test'](c);
	/*return (
		is_c_var_name_char(c, is_start) ||
		c == '-'
	);*/
}

function is_selector_char(c) {
	return /^(\d|\s|\w|_|-|,|\.|#|\*|\[|]|\\|'|"|=|:|>|\||\^|\$|\+|~)/['test'](c);
}

(function (root, factory) {
    if (typeof define === 'function' && define.amd) {
        define(['exports'], factory);
    } else if (typeof exports !== 'undefined') {
        factory(exports);
    } else {
        factory((root.microlight = {}));
    }
}(this, function (exports) {
    // for better compression
    var _window       = window,
        _document     = document,
        appendChild   = 'appendChild',
        test          = 'test',
        // style and color templates
        textShadow    = ';text-shadow:',
        opacity       = 'opacity:.',
        _0px_0px      = ' 0px 0px ',
        _3px_0px_5    = '3px 0px 5',
        brace         = ')',

        is_escaped    = 0,
        is_char       = 0,
        is_str        = 0,
        is_regex      = 0,
        is_at_fn      = 0,
        html_tag_open = 0,
        is_html_tag   = 0,
        is_property   = 0,
        is_selector   = 0,
        is_dollar_var = 0,
        is_tag_name   = 0,
        brace_depth   = 0,
        valid_braces  = 1,
        braces        = [],
        keyword_type  = 0,
        colors = [0, 3, 6, 9, 1, 4, 7, 10, 2, 5, 8],
        colors = ["pink", "orange", "lime", "light-blue", "magenta", "red", "green", "purple", "blue", "indigo"],

        i,
        el,  // current microlighted element to run through
        cl;  // class list
    
    var reset = function(microlighted) {
        // nodes to highlight
        //microlighted = _document.getElementsByClassName(cls||'microlight');
        //microlighted = document.getElementsByClassName(cls||'microlight');

        for (i = 0; el = microlighted[i++];) {
            cl = el.classList;
            is_escaped = is_char
                       = is_str 
                       = is_regex
                       = is_at_fn 
                       = html_tag_open 
                       = is_html_tag 
                       = is_property
                       = is_dollar_var
                       = is_tag_name 
                       = 0;
            valid_braces = 1;
            braces = [];
            brace_depth = 0;
            keyword_type = 0;

            var text  = el.textContent,
                pos   = 0,       // current position
                next1 = text[0], // next character
                chr   = 1,       // current character
                prev1,           // previous character
                prev2,           // the one before the previous
                token =          // current token content
                el.innerHTML = '',  // (and cleaning the node)
                
                // current token type:
                //  0: anything else (whitespaces / newlines)
                //  1: operator or brace
                //  2: closing braces (after which '/' is division not regex)
                //  3: (key)word
                //  4: regex
                //  5: string starting with "
                //  6: string starting with '
                //  7: xml comment  <!-- -->
                //  8: multiline comment /* */
                //  9: single-line comment starting with two slashes //
                // 10: single-line comment starting with hash #
                tokenType = 0,

                // kept to determine between regex and division
                lastTokenType,
                // flag determining if token is multi-character
                multichar,
                node;

            // running through characters and highlighting
            while (prev2 = prev1,
                   // escaping if needed (with except for comments)
                   // pervious character will not be therefore
                   // recognized as a token finalize condition
                   prev1 = tokenType < 7 && prev1 == '\\' ? 1 : chr
            ) {
                chr = next1;
                next1=text[++pos];
                multichar = token.length > 1;

                // checking if current token should be finalized
                if (!chr  || // end of content
                    // types 9-10 (single-line comments) end with a
                    // newline
                    (tokenType > 8 && chr == '\n') ||
                    [ // finalize conditions for other token types
                        // 0: whitespaces
                        prev1 == '$' ||
                        /\S/[test](chr),  // merged together
                        // 1: operators
                        1,                // consist of a single character
                        // 2: braces
                        1,                // consist of a single character
                        // 3: (key)word
                        !/[$\w]/[test](chr) || (/[\d]/[test](prev1) && !/[$\d]/[test](chr)),
                        // 4: regex
                        (prev1 == '/' || prev1 == '\n') && multichar,
                        // 5: string with "
                        prev1 == '"' && multichar,
                        // 6: string with '
                        prev1 == "'" && multichar,
                        // 7: xml comment
                        text[pos-4]+prev2+prev1 == '-->',
                        // 8: multiline comment
                        prev2+prev1 == '*/'
                    ][tokenType]
                ) {
         			// appending the token to the result
                    if (token) {
                        // remapping token type into style
                        // (some types are highlighted similarly)
                        el[appendChild](
                            node = _document.createElement('span')
                        ).setAttribute('style', [
                            // 0: not formatted
                            is_tag_name 
                            	? valid_braces
                            		? "color: var(--neon-" + colors[brace_depth%colors.length] + ")" 
                            		: 'color:var(--primary-dark);background:var(--sorbet-red);padding:1px 0px 2px 0px;-webkit-box-decoration-break:clone;box-decoration-break:clone;'
                            : is_html_tag 
                                ? "color: var(--sherbet-blue)"
                            : is_dollar_var || prev1 == '$' 
                                ? "color:var(--neon-orange)" 
                            : is_at_fn || prev1 == '\@' || chr == '\@'
                                ? "color:var(--neon-pink)" 
                            : chr == '('
                            	? "color:var(--sorbet-orange)"
                            : chr == '['
                            	? "color:var(--sorbet-green)"
                            //inferred types
                            : /^(auto|let|var)$/[test](token)
                            	? "color:var(--sorbet-cyan)"
                            //types
                            : /^(b(ool(|ean))|char|double|int(|eger)|num|s(ize_t|tring))$/[test](token)
                            	? "color:var(--sorbet-light-blue)"
                           	//loops and control statements
                           	: /^(do|el(se|if)|for(|each)|if|switch|while)$/[test](token)
                            	? "color:var(--sorbet-indigo)"
                            //key fns
                           	: /^(main)$/[test](token)
                            	? "color:var(--sorbet-orange)"
                            //(s)css selectors
                            : is_selector || chr == '{' || next1 == '{' || text[pos+1] == '{'
                            	? "color:var(--neon-" + colors[(brace_depth+2)%colors.length] + ")"
                            	//? "color:var(--sorbet-green)"
                            //(s)css properties
                            : chr == ':' || next1 == ':' || text[pos+1] == ':' || is_property
                            	? "color:var(--sorbet-" + colors[(brace_depth+4)%colors.length] + ")"
                            	//? "color:var(--sorbet-green)"
                           	//
                            //: 'color:var(--sherbet-green)',
                            : 'color:var(--sorbet-' + colors[(brace_depth+6)%colors.length] + ')',
                            // 1: keywords
                            'color:var(--sorbet-violet);',
                            // 2: punctuation
                            (brace_depth < 0 || !valid_braces) ? 'color:var(--primary-dark);background:var(--sorbet-red);padding:1px 0px 2px 0px;-webkit-box-decoration-break:clone;box-decoration-break:clone;' :
                            /*prev1 == '@' ? "color:var(--neon-pink)" :*/
                            'color:' + (
                                is_brace(prev1) || 
                                (prev1 == '<' && is_html_tag) || 
                                (prev1 == '>' && is_html_tag) || 
                                (prev1 == '/' && prev2 == '<') ? "var(--neon-" + colors[brace_depth%colors.length] + ")"
                            : prev1 == '-' && chr.match(/^[0-9]+$/) ? "var(--neon-purple)" 
                            : "var(--sorbet-magenta)") + ';',
                            //: "#c75656") + ';',
                            // 3: strings and regexps
                            (is_str || prev1 == '"') ? 'color:var(--sorbet-lime);' : 'color:var(--sherbet-blue);',
                            // 4: numbers
                            'color: var(--sherbet-green)',
                            // 5: comments
                            'color:#9e9e9e;',
                            // 6: types
                            'color: var(--sorbet-pink);'
                        ][
                            // not formatted
                            !tokenType ? 0 :
                            // punctuation
                            tokenType < 3 ? 2 :
                            //types
                            tokenType == 11 ? 6 :
                            // comments
                            tokenType > 6 && tokenType < 11 ? 5 :
                            // regex and strings
                            tokenType > 3 ? 3 :

                            /^\d+$/[test](token) ? 4 :

                            // otherwise tokenType == 3, (key)word
                            // (1 if regexp matches, 0 otherwise)
                            + /^(a(bstract|lias|nd|rguments|rray|s(m|sert)?)|b(ase|egin|reak|yte)|c(ase|atch|hecked|lass|lone|ompl|onst|ontinue|out)|de(bugger|cimal|clare|f(ault)?|init|l(egate|ete)?)|e(cho|nd(l)?|nsure|vent|x(cept|ec|p(licit|ort)|te(nds|nsion|rn)))|f(allthrough|alse|inal(ly)?|ixed|loat|riend|rom|unc(tion)?)|global|goto|guard|i(mp(lements|licit|ort)|n(sert|clude(_once)?|out|stanceof|t(erface|ernal))?|s)|l(ambda|ock|ong)|m(icrolight|odule|utable)|NaN|n(amespace|ative|ext|ew|il|ot|ull)|o(bject|perator|r|ut|verride)|p(ackage|arams|rivate|rotected|rotocol|ublic)|r(aise|e(adonly|do|f|gister|peat|quire(_once)?|scue|strict|try|turn))|s(byte|ealed|elf|hort|igned|izeof|tatic|td|truct|ubscript|uper|ynchronized)|t(emplate|hen|his|hrows?|ransient|rue|ry|ype(alias|def|id|name|of))|u(n(checked|def(ined)?|ion|less|signed|til)|se|sing)|v(ar|irtual|oid|olatile)|w(char_t|hen|here|ith)|xor|yield)$/[test](token)
                        ]);

                        //if(keyword_type)
                        //	alert(keyword_type);

                        //if(tokenType == 3)
                          //alert(token + " " + tokenType);

                        node[appendChild](_document.createTextNode(token));
                    }

                    // saving the previous token type
                    // (skipping whitespaces and comments)
                    lastTokenType =
                        (tokenType && tokenType < 7) ?
                        tokenType : lastTokenType;

                    // initializing a new token
                    token = '';
                    keyword_type = 0;
                    valid_braces = true;

                    // determining the new token type (going up the
                    // list until matching a token type start
                    // condition)
                    tokenType = 11;

                    while (![
                        1,                   //  0: whitespace
                                             //  1: operator or braces
                        /[/{[(%\-+*=<:;|\\.,?!&@~]/[test](chr),
                        /[}\])>]/[test](chr),  //  2: closing brace
                        /[\w]/[test](chr) || /[@\w]/[test](chr),  //  3: (key)word
                        chr == '/' &&  !is_str &&      //  4: regex
                            // previous token was an
                            // opening brace or an
                            // operator (otherwise
                            // division, not a regex)
                            (lastTokenType < 3) &&
                            // workaround for xml
                            // closing tags
                            prev1 != '<' && prev1 != '\\',
                        chr == '"',          //  5: string with "
                        chr == "'",          //  6: string with '
                                             //  7: xml comment
                        chr+next1+text[pos+1]+text[pos+2] == '<!--',
                        chr+next1 == '/*',   //  8: multiline comment
                        chr+next1 == '//',   //  9: single-line comment
                        chr == '#'          // 10: hash-style comment
                    ][--tokenType]);
                }

                if(!is_str && !is_regex && is_open_brace(prev1)) {
                    ++brace_depth;
                    braces.push(prev1);
                }

                if(!is_str && !is_regex && is_close_brace(chr)) {
                	--brace_depth;
                    valid_braces = (
                    	(chr == '}' && braces[brace_depth] == '{') ||
                    	(chr == ']' && braces[brace_depth] == '[') ||
                    	(chr == ')' && braces[brace_depth] == '(') 
                    );
                    //if(!valid_braces)
                    //	alert(brace_depth + " " + chr + " " + braces[brace_depth] + " " + text.substr(pos, 10));
                    braces.pop();
                }

                //is_escaped = (!is_escaped && prev1 == '\\');
                if(is_escaped)
                	is_escaped = 0;
                else if(prev1 == '\\')
                	is_escaped = 1;

                if(!is_escaped && !is_str && chr == "'")
                	is_char = !is_char;

                if(!is_escaped && !is_char && chr == '"')
                	is_str = !is_str;

                if(!is_str && chr == '/' && (lastTokenType < 3 || is_regex) && prev1 != '<')
                	is_regex = !is_regex;

                if(prev1 == '\@' && !is_whitespace(chr) && !is_brace(chr))
                    is_at_fn = 1;
                else if(is_at_fn && (is_whitespace(chr) || is_brace(chr)))
                    is_at_fn = 0;

                if(prev1 == '$' && !is_whitespace(chr) && !is_brace(chr) && !('0' <= chr && chr <= '9'))
                    is_dollar_var = 1;
                else if(is_dollar_var && (is_whitespace(chr) || is_brace(chr)))
                    is_dollar_var = 0;

                if(is_tag_name && (is_whitespace(chr) || is_brace(chr)))
                    is_tag_name = 0;

                if(is_selector && next1 == '{')
                	is_selector = 0;
                else if(!is_selector && /^(\s|,|\.|:|-|>|~|\+)/['test'](next1)) {
                	for(let p=pos+1; p<text.length; ++p) {
                		if(text[p] == '{') {
                			is_selector = 1;
                			break;
                		}
                		else if(!is_selector_char(text[p])) {
                			break;
                		}
                	}
                }

                if(is_property && next1 == ':')
                	is_property = 0;
                else if(!is_property && next1 == "-") {
                	for(let p=pos+1; p<text.length; ++p) {
                		if(text[p] == ':') {
                			is_property = 1;
                			break;
                		}
                		//else if(/^(\S|\W|-)/[test](text[p])) //test this
                		else if(
                			!('a' <= text[p] && text[p] <= 'z') &&
                        	!('A' <= text[p] && text[p] <= 'Z') &&
                        	!is_whitespace(text[p]) &&
                        	text[p] != '-'
                		)
                			break;
                	}
                }

                if(is_html_tag && prev1 == '>') {
                    is_html_tag = 0;
                    is_tag_name = 0;
                    if(html_tag_open) {
                        ++brace_depth;
                        braces.push('<');
                    }
                }


                if(chr == '<') {
                    if(next1 == '/'){
                        is_html_tag = 1;
                        is_tag_name = 1;
                        html_tag_open = 0; 
                        --brace_depth;
                        valid_braces = (braces[brace_depth] == '<');
	                    braces.pop();
                    }
                    else if(/^(\w|!)/[test](next1)) {
                        let depth = 0;
                        let x = pos+1;

                        for(; x<text.length; ++x) {
                            if(!depth && text[x] == '>') {
                                is_html_tag = 1;
                                html_tag_open = 1;
                                is_tag_name = 1;
                                break;
                            }
                            else if(/^("|')/[test](text[x])) {
                                let end_quote = text[x];
                                while(
                                    ++x < text.length && 
                                    text[x] != end_quote
                                )
                                    if(text[x] == '\\')
                                        ++x;
                            }
                            else if(/^(;|}|])/[test](text[x]))
                                break;
                            else if(/^({|<|\[)/[test](text[x]))
                                ++depth;
                            else if(/^(}|>|])/[test](text[x]))
                                --depth;
                        }

                        if(x == text.length) {
                        	is_html_tag = 1;
                            html_tag_open = 1;
                            is_tag_name = 1;
                        }
                    }
                }

                token += chr;
            }
        }
    }

    var reset_id = function(id) {
    	reset([_document.getElementById(id)]);
    }

    exports.reset = reset;
    exports.reset_id = reset_id;

    if (_document.readyState == 'complete') {
        reset(_document.getElementsByClassName("microlight"));
    } else {
        _window.addEventListener('load', function(){reset(_document.getElementsByClassName("microlight"))}, 0);
    }
}));
