interface User {
  email: string,
  address: Address,
  phone: string,
  website: string,
  company: Company,
  id: number,
  name: string,
  username: string,
}
interface Company {
  bs: string,
  name: string,
  catchPhrase: string,
}
interface Address {
  zipcode: string,
  geo: Geo,
  street: string,
  suite: string,
  city: string,
}
interface Geo {
  lat: string,
  lng: string,
}
