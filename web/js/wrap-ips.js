document.querySelectorAll('.ip-address').forEach(el => {
  el.innerHTML = el.textContent.replace(/[:.]/g, '$&<wbr>'); //inserts an invisible word break character after each colon and period. 
});                                                      //this character is a valid break, IP address will wrap only here 
