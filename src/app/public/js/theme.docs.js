/** 

  Anchor JS
  Injected links

  <h2 class="doc-anchor">
    Lorem Ipsum
  </h2>

**/

if( window.location.origin == 'file://' ) {
  alert('Please, use a local webserver!')
}

function docAnchor() {

  let anchNodes = document.querySelectorAll('.doc-anchor');
  if( !anchNodes ) return;

  anchNodes.forEach(function(el) {

    // ignore multiple bind -- -- -- -- -- -- -- -- --
    if( el.classList.contains('anchor-js') ) return;
      el.classList.add('anchor-js');
    // -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --

    let anchor = el.innerHTML;
    if( !anchor ) anchor = Math.random().toString(36).substring(7);

    // stip tags
    let divEl = document.createElement("div");
        divEl.innerHTML = anchor;
        anchor = divEl.textContent || divEl.innerText || anchor;

    anchor = anchor.toLowerCase().trim();
    anchor = anchor.toString().replace(/\./g, '_')
                              .replace(/\s+/g, '-')
                              .replace(/[^a-zA-Z0-9-_ ]/g, '')
                              .replace(/---/g, '-')
                              .replace(/--/g, '-');

    el.setAttribute('id', anchor);
    el.insertAdjacentHTML( 'beforeend', '<a class="anchor-js-link mx-2" href="#'+anchor+'" aria-label="Anchor"></a>' );

  });

} docAnchor();



/**
  
  Selected navigation
  accoding to current url

**/
function docNavSelected() {

  // url  : https://localhost/smarty/html_admin/layout_2/index.html
  // html : index.html
  let url       = window.location.href;
  let urlSplit  = url.split('/');
  let html      = urlSplit[urlSplit.length - 1];
      html      = html.split('#');
      html      = html[0];

  // mark active
  let asideHref = document.querySelector('aside a.js-ajax[href="'+html+'"]');
  if( asideHref ) {

    asideHref.parentElement.classList.add('active');

    // -- --

    // parent #1 (deep nav)
    let parentUL = asideHref.closest('ul');
    if( parentUL ) {

      // mark active
      parentUL.parentElement.classList.add('active');

      // parent #2 (deep nav)
      let parentUL2 = parentUL.closest('ul');
      if( parentUL2 ) parentUL2.parentElement.classList.add('active');

    }

  }

} docNavSelected();
