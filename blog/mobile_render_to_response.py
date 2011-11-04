from django.shortcuts import render_to_response as render_to_response_pre
from django.template import RequestContext

MOBILE_USERAGENTS = ("2.0 MMP","240x320","400X240","AvantGo","BlackBerry",
    "Blazer","Cellphone","Danger","DoCoMo","Elaine/3.0","EudoraWeb",
    "Googlebot-Mobile","hiptop","IEMobile","KYOCERA/WX310K","LG/U990",
    "MIDP-2.","MMEF20","MOT-V","NetFront","Newt","Nintendo Wii","Nitro",
    "Nokia","Opera Mini","Palm","PlayStation Portable","portalmmm","Proxinet",
    "ProxiNet","SHARP-TQ-GX10","SHG-i900","Small","SonyEricsson","Symbian OS",
    "SymbianOS","TS21i-10","UP.Browser","UP.Link","webOS","Windows CE",
    "WinWAP","YahooSeeker/M1A1-R2D2","iPhone","iPod","Android",
    "BlackBerry9530","LG-TU915 Obigo","LGE VX","webOS","Nokia5800")

def render_to_response(template,dictionary,context):
	if any(ua for ua in MOBILE_USERAGENTS if ua in context['request'].META.get("HTTP_USER_AGENT","")):
		template = template.rstrip('.html')
		template += '.mobile.html'
	return render_to_response_pre(template,dictionary, context)

