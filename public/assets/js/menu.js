const menu_page = document.getElementById("menu-page");
const content_page = document.getElementById("content");
const menu_bars_icon = document.getElementById("open-menu-icon");

function show_menu() {
  menu_page.style.display = "block";
  content_page.style.display = "none";
  menu_bars_icon.style.display = "none";
}

function close_menu() {
  menu_page.style.display = "none";
  content_page.style.display = "block";
  menu_bars_icon.style.display = "block";
}

function toggle_submenu(i) {
  if(!window.event.ctrlKey) {
    item = document.getElementsByClassName('has-submenu')[i];
    item_link = document.getElementsByClassName('submenu-link')[i];
    submenu = document.getElementsByClassName('submenu')[i];
    if(item.classList.contains("submenu-active")) {
      item.classList.remove("submenu-active");
      item_link.classList.remove("submenu-link-active");
      submenu.style.display = "none";
    } 
    else {
      item.classList.add("submenu-active");
      item_link.classList.add("submenu-link-active");
      submenu.style.display = "flex";
    }
  }
}
