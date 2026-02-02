# K.I.S.S principle
    ^
    |
 **Keep It Simple Stupid**

 All endpoints are in ``` routes.go``` file

 1.) -> First page is *login page* ```go public.POST("/login", authHandler.Login) ```. Tenants have no         dashbords or accounts to the system at the moment(reason we have few landlords :).
        we are going to make use of *google AOuth2*. landlords/Admins will just have to login via google -Everyone uses google, no doubt
                    -Why this helps:
                                a) password remembering for landlords is difficult; strong passwd I mean
                                b) It will do away with password reset functionality; which is always hectic to most principle
                                c) 

2.) -> Admins logs in, he's able to Create user(landlord) etc

    ```go
    adminRoutes := r.Group("/sudo")
	adminRoutes.Use(
		middleware.AuthMiddleware([]byte(cfg.JWT.Secret)), //sets user in context
		middleware.RequireRole(db, "admin"),               //checks role
	)
	{
		adminRoutes.POST("register", authHandler.Register) //register landlord
		//admin.GET("/users", handlers.ListLandlords)
 		//we'll add more functionality as the app grows(K.I.S.S)
	}
   ```
   **Reqire role** function checks in users table under role column if role matches

3.) -> Landlord logs in:

```go 
landlord := r.Group("/api/v1")
	landlord.Use(middleware.AuthMiddleware([]byte(cfg.JWT.Secret)))
	{
		// Properties
		landlord.POST("/properties/newProperty", handlers.CreateProperty(db))
		landlord.GET("/properties/myProperties", handlers.ListProperties(db))
		landlord.GET("/properties/:propertyId", handlers.GetProperty(db))
		landlord.PATCH("/properties/:updatePropertyByID", handlers.UpdateProperty(db))
		//landlord.DELETE("/properties/:propertyId", controllers.DeleteProperty) // they can't  delete property at the moment,

		// Units (nested under properties)
		landlord.POST("/properties/:propertyId/newUnit", handlers.CreateUnit(db))
		landlord.GET("/properties/:propertyId/units", handlers.GetUnitsByProperty(db))
		//landlord.GET("/units/:unitId", controllers.GetUnit) --jj
	    //landlord.PUT("/units/:unitId", controllers.UpdateUnit)
		//landlord.DELETE("/units/:unitId", controllers.DeleteUnit) --ii

		// Tenants (assigned to units)
		landlord.POST("/units/:unitId/addTenant", handlers.CreateTenant(db))
		landlord.GET("/units/:unitId/tenants", handlers.ListTenants(db))
		landlord.GET("/tenants/:tenantId", handlers.GetTenant(db))
		landlord.PUT("/tenants/:tenantId", handlers.UpdateTenant(db))
		landlord.DELETE("/tenants/:tenantId", handlers.RemoveTenant(db))
	}
```

I've used jwt as http only cookie to set user id in context and not accept user id from request body ie
 ```json "landlord_id: 1"```; frontend does not interact with user id, browser fetches id directly from the server.This avoids CSRF

Tenant pays normaly from his phone, since we have have his payment no in the database, it will automatically reflect as paid and show balance, just like school fee. Customer to Business(C2B) daraja api will apply here

I'll add google auth2.0 last, plus calling my rate limiter function

   
    

    