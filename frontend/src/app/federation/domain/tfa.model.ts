export interface UserResponse {
	id :number
	login: string
	email: string
	firstName: string
	lastName: string
	imageUrl: string
	langKey: string
	tfaEnabled: boolean
	tfaMethod: string
}
