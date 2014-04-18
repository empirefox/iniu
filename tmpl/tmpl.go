package tmpl

var bucketsHtml = `<!DOCTYPE html><html>
<head>
    <meta charset="utf-8"></meta>
    <link rel="stylesheet" type="text/css" href="//cdn.staticfile.org/twitter-bootstrap/3.1.1/css/bootstrap.min.css"></link>
    <script type="text/javascript" src="//cdn.staticfile.org/angular.js/1.0.8/angular.min.js"></script>
    <script type="text/javascript" src="//cdn.staticfile.org/angular-ui-bootstrap/0.10.0/ui-bootstrap.min.js"></script>
</head>
<body>
<div ng-app="myApp" class="container">
	<div><a href="/logout">退出</a><div>
	<div ng-controller="BucketsCtrl" class="row">
		<div class="col-md-3">
			<div class="list-group">
				<a ng-click="edit({Name:'新名称',Life:380})" class="list-group-item" href="#">新建</a>
				<a ng-repeat="b in buckets" ng-click="edit(b)" class="list-group-item" href="#">{{b.Name}}</a>
			</div>
		</div>
		<div class="col-md-9">
		<form action="/bucket" method="post" class="form-horizontal" role="form">
		[[range .Fields]]
			[[.|toControl]]
		[[end]]
		<div class="form-group">
			<div class="col-md-offset-2 col-md-10">
				<button type="submit" class="btn btn-default">提交</button>
				<button ng-click="remove()" type="button" class="btn btn-default">删除</button>
			</div>
		</div>
		</form>
		</div>
	</div>
	<script>
		Array.prototype.remove = function(val) {  
		var index = this.indexOf(val);
			console.log('删除>')
			if (index > -1) {
				this.splice(index, 1);
				console.log('成功:'+index)
			}
		};
		var BucketsCtrl = function($scope,$http){
			$scope.buckets = [[.Buckets|json]];
			$scope.bucket = $scope.buckets[0]
			$scope.edit = function(b){
				$scope.bucket = b
			}
			$scope.remove = function(){
				$http.post('/remove_bucket',{Id:$scope.bucket.Id}).success(function(data, status){
					console.log(data)
					if(data.error===0){
						$scope.buckets.remove($scope.bucket)
					}
				});
			}
		};
		var app = angular.module('myApp',['ui.bootstrap'],function(){});
 		//angular.bootstrap(document, ['myApp']);
	</script>
</div> 
</body>
</html>`
