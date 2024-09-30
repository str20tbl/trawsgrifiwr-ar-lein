/** 

  Documentation : Icons

  <ul class="list doc-icon-preview" data-vendor-path="../assets/svg/icons/feather/" data-vendor-available="img,svg,webfont" data-vendor-webfont-prefix="fe fe-">
    <li data-icon="icon-name">
      <span class="svg-wrapper">
        ... svg ...
      </span>
      <p>icon-name</p>
    </li>
  </ul>

**/
document.querySelectorAll('.doc-icon-preview').forEach(function(_t) {

  var vendor_path             = _t.getAttribute('data-vendor-path')           || 'assets/svg/icons/';
  var vendor_webfont_prefix   = _t.getAttribute('data-vendor-webfont-prefix') || 'fi fi-';
  var vendor_available_list   = _t.getAttribute('data-vendor-available')      || 'img,svg,webfont';
      vendor_available_list   = vendor_available_list.replace(/ /gi, '');
      vendor_available_list   = vendor_available_list.split(',');

  console.log('Total Icons:', _t.querySelectorAll('li').length );

  _t.querySelectorAll('li').forEach(function(el) {

    // ignore multiple bind -- -- -- -- -- -- -- -- --
    if( el.classList.contains('icon-js') ) return;
      el.classList.add('icon-js');
    // -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --

    let svgItem = el.querySelector('.svg-wrapper>svg');
      if( svgItem ) {
        // svgItem.removeAttribute('width');        // overwrite
        svgItem.setAttribute('height', '1.5em');    // overwrite
        svgItem.setAttribute('width', '1.5em');     // overwrite
      }

    let icon_name = el.getAttribute('data-icon');


    // buttons
    let _btnTpl = '<span class="icon-opt animate-slidein">';
    // -- -- -- 
    if( vendor_available_list.indexOf('webfont') !== -1 )
      _btnTpl += '<a class="btn btn-sm btn-warning shadow-xl-warning btn-clipboard" href="#" data-icon="'+icon_name+'" data-icon-prefix="'+vendor_webfont_prefix+'">ICO</a>';

    if( vendor_available_list.indexOf('svg') !== -1 )
      _btnTpl += '<a class="btn btn-sm btn-warning shadow-xl-warning btn-opt-icon icon-svg" href="#modal-icon" data-bs-toggle="modal">SVG</a>';

    if( vendor_available_list.indexOf('img') !== -1 )
      _btnTpl += '<a class="btn btn-sm btn-warning shadow-xl-warning btn-opt-icon icon-img" href="#modal-icon" data-bs-toggle="modal">IMG</a>';
    // -- -- -- 
    _btnTpl += '</span>';
    el.insertAdjacentHTML( 'beforeend', _btnTpl );



    // bind click
    el.querySelectorAll('.btn-opt-icon').forEach(function(btn) {

      btn.addEventListener('click', function(e) {
        e.preventDefault();

        let icon_img  = '<img width="18" height="18" src="'+vendor_path+icon_name+'.svg" alt="'+icon_name+'">';
        let icon_svg  = el.querySelector('.svg-wrapper').innerHTML;


        // Image
        if( ( e.target.classList.contains('icon-img') ) ) {

          var icon = icon_img.trim().replace(/\</gi, '&lt;').replace(/\>/gi, '&gt;');

        // SVG
        } else {

          var icon = icon_svg.trim()
                            .replace(/\>/gi, '&gt;')
                            .replace('<svg', '&lt;svg')
                            .replace('</svg', '\n&lt;/svg')
                            .replace(/\<\//gi, "&lt;/")
                            .replace(/\</gi, "\n  &lt;")
                            .replace(/1em/gi, "18px").replace(/1.5em/gi, "18px");

        }


        document.querySelector('#mdlTitle').innerHTML = icon_svg.replace('svg ', 'svg class="float-start me-3 text-primary" ') + icon_name;
        document.querySelector('#modal-icon .modal-body>pre').innerHTML = icon;

      });

    });

  });
});






/* 

  Clipboard 
  <li class="btn-clipboard" data-icon="icon-name"> ... </li>
  <a href="#!" class="btn-clipboard" data-icon="icon-name"> ... </a>
  
  <script src="../assets/js/vendor.clipboard.min.js"></script>

*/
if( typeof ClipboardJS === 'function' ) {

  document.body.insertAdjacentHTML('beforeend', '<textarea id="clipboardCode" style="position:fixed;left:-500px;bottom:-500px;height:1px" rows="1"></textarea>');
  var clipboard = new ClipboardJS('.btn-clipboard', {
    target: function (el) {

      let data_icon_prefix = el.getAttribute('data-icon-prefix');
      let data_icon = el.getAttribute('data-icon');
      let final_icon = data_icon_prefix + data_icon;
      
      // -- -- --
      // fix 'fi fi-mdi-icon' => 'fi mdi-icon'
      final_icon = final_icon.replace('fi fi-mdi-', 'fi mdi-');
      // alert(final_icon)
      // -- -- --

              document.getElementById('clipboardCode').value = '<i class="' + final_icon + '"></i>';
      return  document.getElementById('clipboardCode');

    }
  });

  clipboard.on('success', function (e) {
    
    var clipboardTooltip = '<div id="clipboardTooltip" class="p-3 mb-1 text-center shadow-md bg-warning animate-slidein" style="width:200px;left:50%;margin-left:-100px;position:fixed;bottom:0;border-radius:.25rem;">Copied to clipboard</div>';
    document.body.insertAdjacentHTML('beforeend',clipboardTooltip);

    setTimeout(function () {
    
      var clipboardTooltipEl = document.getElementById('clipboardTooltip');
      clipboardTooltipEl.parentNode.removeChild(clipboardTooltipEl);
    
    }, 1500);

  });

  clipboard.on('error', function () {

    var clipboardTooltip = '<div id="clipboardTooltip" class="p-3 mb-1 text-center shadow-md bg-danger animate-slidein" style="width:120px;left:50%;margin-left:-60px;position:fixed;bottom:0;border-radius:.25rem;">ERROR</div>';
    document.body.insertAdjacentHTML('beforeend',clipboardTooltip);

    setTimeout(function () {
    
      var clipboardTooltipEl = document.getElementById('clipboardTooltip');
      clipboardTooltipEl.parentNode.removeChild(clipboardTooltipEl);
    
    }, 2000);

  });

  document.querySelectorAll('.btn-clipboard').forEach(function(item) {
    item.addEventListener('click', function(e) {
      e.preventDefault();
    });
  });

}
