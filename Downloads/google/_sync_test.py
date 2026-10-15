"""Run uvicorn synchronously to see real errors."""
import sys
import traceback

sys.path.insert(0, "/Users/bikashtalukder/Downloads/google/_minimal_api.py")

try:
    import _minimal_api as m
    print("Module:", m)
    print("App:", m.app)

    # Try to call the handler directly
    from starlette.requests import Request
    from starlette.responses import Response

    # Simulate a request
    scope = {"type": "http", "method": "GET", "path": "/healthz", "headers": []}
    req = Request(scope)

    # Get the route
    routes = m.app.router.routes
    for r in routes:
        print("Route:", r)
        if r.path == "/healthz":
            try:
                resp = r.endpoint()
                print("Direct call OK:", resp)
            except Exception as e:
                print("Direct call FAILED:", e)
                traceback.print_exc()
            break
except Exception as e:
    print("Module load FAILED:", e)
    traceback.print_exc()
