package Utils




func ArrCompare(slice []string , item string) bool{
  for _, i := range slice{
    if i == item{
    return true
    }
  }
  return false
}
