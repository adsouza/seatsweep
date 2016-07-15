window.addEventListener('load', function(){

    window.clicks = [];    

    var map = document.getElementById('map');
    
    var canvas = document.getElementById("canvas");
    canvas.setAttribute('width', map.naturalWidth);     
    canvas.setAttribute('height', map.naturalHeight);

    var heatmap_canvas = document.getElementById("heatmap");
    heatmap_canvas.setAttribute('width', map.naturalWidth);     
    heatmap_canvas.setAttribute('height', map.naturalHeight);
    window.heat = simpleheat(heatmap_canvas);
    heat.resize();

    var map = document.getElementById('map');
    map.addEventListener('click', processAddPoint);

    var makeheatmapbutton = document.getElementById('makeheatmap');
    makeheatmapbutton.addEventListener('click', renderHeatMap);

}, false);

function processAddPoint(event){
    var map = document.getElementById('map');
    var canvas = document.getElementById("canvas");

    var xAxisRatio = canvas.getAttribute('width') / map.width;
    var yAxisRatio = canvas.getAttribute('height') / map.height;
    translatedX = Math.round(event.offsetX * xAxisRatio);
    translatedY = Math.round(event.offsetY * yAxisRatio);

    clicks.push([translatedX, translatedY, 0.5]);

    var ctx = canvas.getContext("2d");
    ctx.fillStyle = 'darkred';
    ctx.beginPath();
    ctx.arc(translatedX, translatedY, 5, 0, 2 * Math.PI);
    ctx.fill();
}

function renderHeatMap(event){
    var canvas = document.getElementById("canvas");
    var ctx = canvas.getContext("2d");
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    heat.data(clicks).draw();
}