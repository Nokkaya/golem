import urllib.request, json
url = 'https://api.open-meteo.com/v1/forecast?latitude=22.5431&longitude=114.0579&current_weather=true&timezone=Asia/Shanghai'
data = json.loads(urllib.request.urlopen(url).read().decode())
w = data['current_weather']
print(f"深圳天气: {w['temperature']}°C, 风速 {w['windspeed']}km/h")