function Api(url, error) {
	var _this = this;
	var _url = url;
	var _type = "POST";
	var _error = error;

	_this.call = function(name, args, invoke, error) {
		if (typeof args == "undefined" || args == null)
			var args = {};
		var error_handler = _error;
		if (typeof error == "function") {
			error_handler = function(data, name, args) {
				_error(data, name, args)
				error(data, name, args);
			}
		}
		_this.poller = $.ajax({
			url: _url + "/" + name,
			type: _type,
			data: args,
			error: function(r, s, e) {
				error_handler("" + r.status + ":" + r.responseText, name, args);
			},
			timeout: 3000,
			success: function(json) {
				if (json == null) {
					error_handler(data, name, args);
					return
				}
				if (json["status"] != "OK") {
					error_handler(data, name, args);
					return
				}
				var result = json["result"];
				if (!result.length || result.length < 1 || result[result.length - 1] != null) {
					error_handler(data, name, args);
					return
				}
				result = result.splice(0, result.length - 1);
				if (result.length == 1) {
					result = result[0];
				}
				if (typeof invoke != "undefined") {
					invoke(result);
				}
			}
		});
	}
}
