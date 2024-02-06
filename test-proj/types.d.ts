interface User {
  name: string,
  username: string,
  email: string,
  address: Address,
  phone: string,
  website: string,
  company: Company,
  id: number,
}
interface Company {
  name: string,
  catchPhrase: string,
  bs: string,
}
interface Address {
  suite: string,
  city: string,
  zipcode: string,
  geo: Geo,
  street: string,
}
interface Geo {
  lng: string,
  lat: string,
}
