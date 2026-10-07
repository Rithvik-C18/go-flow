"""Exercise a disposable local/test deployment; creates two test users."""
import json
import os
import time
import urllib.error
import urllib.request

base = os.environ.get('SMOKE_URL', 'http://localhost:18080').rstrip('/')
suffix = str(time.time_ns())

def request(method, path, data=None, token=None, cookie=None, expected=200):
    headers = {'Content-Type': 'application/json'}
    if token:
        headers['Authorization'] = 'Bearer ' + token
    if cookie:
        headers['Cookie'] = cookie
    req = urllib.request.Request(base + path, data=json.dumps(data).encode() if data is not None else None, headers=headers, method=method)
    try:
        response = urllib.request.urlopen(req, timeout=45)
    except urllib.error.HTTPError as error:
        response = error
    body = response.read()
    assert response.status == expected, (method, path, response.status, body[:300])
    try:
        value = json.loads(body)
    except ValueError:
        value = body.decode()
    return value, response.headers

assert request('GET', '/healthz')[0]['status'] == 'ok'
assert '<title>Fluxion</title>' in request('GET', '/auth')[0]
assert '<title>Fluxion</title>' in request('GET', '/workflows/deep-link')[0]
request('GET', '/api/not-found', expected=404)
request('GET', '/api/workflows', expected=401)
users = []
for index in range(2):
    username = 'smoke_' + suffix + '_' + str(index)
    user, headers = request('POST', '/api/auth/register', {'username': username, 'email': username + '@example.com', 'password': 'SmokeTestOnly!123'}, expected=201)
    cookie = headers['Set-Cookie'].split(';')[0]
    assert 'HttpOnly' in headers['Set-Cookie'] and 'Secure' in headers['Set-Cookie']
    users.append((user['access_token'], cookie))
owner, cookie = users[0]
other = users[1][0]
workflow = 'smoke_' + suffix
request('POST', '/api/workflows', {'id': workflow, 'name': 'Deployment check'}, owner, expected=201)
assert any(w['id'] == workflow for w in request('GET', '/api/workflows', token=owner)[0])
assert not any(w['id'] == workflow for w in request('GET', '/api/workflows', token=other)[0])
for method, path, data in [
    ('GET', '', None), ('DELETE', '', None), ('POST', '/run', {}),
    ('GET', '/nodes', None), ('POST', '/nodes', {'id':'x', 'type':'if'}),
    ('DELETE', '/nodes/x', None), ('POST', '/edges', {'from':'x','to':'y'}),
    ('DELETE', '/edges', {'from':'x','to':'y'}),
]:
    request(method, '/api/workflows/' + workflow + path, data, other, expected=404)
node = {'id': 'condition', 'type': 'if', 'config': {'conditions':[{'left':1,'operator':'==','right':1}]}}
request('POST', '/api/workflows/' + workflow + '/nodes', node, owner, expected=201)
assert request('POST', '/api/workflows/' + workflow + '/run', {}, owner)[0]['status'] == 'completed'
request('DELETE', '/api/workflows/' + workflow + '/nodes/condition', token=owner)
request('DELETE', '/api/workflows/' + workflow, token=owner)
request('GET', '/api/workflows/' + workflow, token=owner, expected=404)
refreshed, headers = request('POST', '/api/refresh', cookie=cookie)
cookie = headers['Set-Cookie'].split(';')[0]
request('POST', '/api/logout', token=refreshed['access_token'])
request('POST', '/api/refresh', cookie=cookie, expected=401)
print('PASS: health, frontend/deep links, registration, workflow CRUD/execution, cross-user isolation, refresh and logout')
