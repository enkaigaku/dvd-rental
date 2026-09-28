// gen_proto_rawdesc generates rawDesc bytes for payment and customer proto files.
// Run with: go run ./tools/gen_proto_rawdesc/main.go
package main

import (
	"bytes"
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
)

func pstr(s string) *string { return &s }
func pint32(i int32) *int32 { return &i }

func fieldType(t descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto_Type {
	return &t
}
func fieldLabel(l descriptorpb.FieldDescriptorProto_Label) *descriptorpb.FieldDescriptorProto_Label {
	return &l
}

func int32Field(name string, num int32) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:   pstr(name),
		Number: pint32(num),
		Type:   fieldType(descriptorpb.FieldDescriptorProto_TYPE_INT32),
		Label:  fieldLabel(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
	}
}

func boolField(name string, num int32) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:   pstr(name),
		Number: pint32(num),
		Type:   fieldType(descriptorpb.FieldDescriptorProto_TYPE_BOOL),
		Label:  fieldLabel(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
	}
}

func stringField(name string, num int32) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:   pstr(name),
		Number: pint32(num),
		Type:   fieldType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
		Label:  fieldLabel(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
	}
}

func messageField(name string, num int32, typeName string) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:     pstr(name),
		Number:   pint32(num),
		Type:     fieldType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
		Label:    fieldLabel(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL),
		TypeName: pstr(typeName),
	}
}

func repeatedMessageField(name string, num int32, typeName string) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:     pstr(name),
		Number:   pint32(num),
		Type:     fieldType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
		Label:    fieldLabel(descriptorpb.FieldDescriptorProto_LABEL_REPEATED),
		TypeName: pstr(typeName),
	}
}

func repeatedStringField(name string, num int32) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:   pstr(name),
		Number: pint32(num),
		Type:   fieldType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
		Label:  fieldLabel(descriptorpb.FieldDescriptorProto_LABEL_REPEATED),
	}
}

func msg(name string, fields ...*descriptorpb.FieldDescriptorProto) *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name:  pstr(name),
		Field: fields,
	}
}

func rpc(name, inputType, outputType string) *descriptorpb.MethodDescriptorProto {
	return &descriptorpb.MethodDescriptorProto{
		Name:       pstr(name),
		InputType:  pstr(inputType),
		OutputType: pstr(outputType),
	}
}

func svc(name string, methods ...*descriptorpb.MethodDescriptorProto) *descriptorpb.ServiceDescriptorProto {
	return &descriptorpb.ServiceDescriptorProto{
		Name:   pstr(name),
		Method: methods,
	}
}

func encodeRawDesc(fdp *descriptorpb.FileDescriptorProto) string {
	b, err := proto.Marshal(fdp)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	buf.WriteString("\t\"")
	for i, c := range b {
		if i > 0 && i%16 == 0 {
			buf.WriteString("\" +\n\t\"")
		}
		buf.WriteString(fmt.Sprintf("\\x%02x", c))
	}
	buf.WriteString("\"")
	return buf.String()
}

func paymentProto() *descriptorpb.FileDescriptorProto {
	return &descriptorpb.FileDescriptorProto{
		Name:    pstr("payment/v1/payment.proto"),
		Package: pstr("payment.v1"),
		Options: &descriptorpb.FileOptions{
			GoPackage: pstr("github.com/enkaigaku/dvd-rental/gen/proto/payment/v1;paymentv1"),
		},
		Syntax:     pstr("proto3"),
		Dependency: []string{"google/protobuf/empty.proto", "google/protobuf/timestamp.proto"},
		MessageType: []*descriptorpb.DescriptorProto{
			msg("Payment",
				int32Field("payment_id", 1),
				int32Field("customer_id", 2),
				int32Field("staff_id", 3),
				int32Field("rental_id", 4),
				stringField("amount", 5),
				messageField("payment_date", 6, ".google.protobuf.Timestamp"),
			),
			msg("PaymentDetail",
				messageField("payment", 1, ".payment.v1.Payment"),
				stringField("customer_name", 2),
				stringField("staff_name", 3),
				messageField("rental_date", 4, ".google.protobuf.Timestamp"),
			),
			msg("GetPaymentRequest",
				int32Field("payment_id", 1),
			),
			msg("ListPaymentsRequest",
				int32Field("page_size", 1),
				int32Field("page", 2),
			),
			msg("ListPaymentsResponse",
				repeatedMessageField("payments", 1, ".payment.v1.Payment"),
				int32Field("total_count", 2),
			),
			msg("ListPaymentsByCustomerRequest",
				int32Field("customer_id", 1),
				int32Field("page_size", 2),
				int32Field("page", 3),
			),
			msg("ListPaymentsByStaffRequest",
				int32Field("staff_id", 1),
				int32Field("page_size", 2),
				int32Field("page", 3),
			),
			msg("ListPaymentsByRentalRequest",
				int32Field("rental_id", 1),
				int32Field("page_size", 2),
				int32Field("page", 3),
			),
			msg("ListPaymentsByDateRangeRequest",
				messageField("start_date", 1, ".google.protobuf.Timestamp"),
				messageField("end_date", 2, ".google.protobuf.Timestamp"),
				int32Field("page_size", 3),
				int32Field("page", 4),
			),
			msg("CreatePaymentRequest",
				int32Field("customer_id", 1),
				int32Field("staff_id", 2),
				int32Field("rental_id", 3),
				stringField("amount", 4),
			),
			msg("DeletePaymentRequest",
				int32Field("payment_id", 1),
			),
			msg("GetCustomerBalanceRequest",
				int32Field("customer_id", 1),
			),
			msg("CustomerBalance",
				int32Field("customer_id", 1),
				stringField("total_charges", 2),
				stringField("total_payments", 3),
				stringField("balance", 4),
				int32Field("rental_count", 5),
				int32Field("payment_count", 6),
			),
			msg("GetRevenueByStoreRequest",
				messageField("start_date", 1, ".google.protobuf.Timestamp"),
				messageField("end_date", 2, ".google.protobuf.Timestamp"),
			),
			msg("StoreRevenue",
				int32Field("store_id", 1),
				stringField("total_revenue", 2),
				int32Field("payment_count", 3),
				int32Field("rental_count", 4),
			),
			msg("RevenueByStoreResponse",
				repeatedMessageField("stores", 1, ".payment.v1.StoreRevenue"),
				stringField("total_revenue", 2),
			),
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			svc("PaymentService",
				rpc("GetPayment", ".payment.v1.GetPaymentRequest", ".payment.v1.PaymentDetail"),
				rpc("ListPayments", ".payment.v1.ListPaymentsRequest", ".payment.v1.ListPaymentsResponse"),
				rpc("ListPaymentsByCustomer", ".payment.v1.ListPaymentsByCustomerRequest", ".payment.v1.ListPaymentsResponse"),
				rpc("ListPaymentsByStaff", ".payment.v1.ListPaymentsByStaffRequest", ".payment.v1.ListPaymentsResponse"),
				rpc("ListPaymentsByRental", ".payment.v1.ListPaymentsByRentalRequest", ".payment.v1.ListPaymentsResponse"),
				rpc("ListPaymentsByDateRange", ".payment.v1.ListPaymentsByDateRangeRequest", ".payment.v1.ListPaymentsResponse"),
				rpc("CreatePayment", ".payment.v1.CreatePaymentRequest", ".payment.v1.Payment"),
				rpc("DeletePayment", ".payment.v1.DeletePaymentRequest", ".google.protobuf.Empty"),
				rpc("GetCustomerBalance", ".payment.v1.GetCustomerBalanceRequest", ".payment.v1.CustomerBalance"),
				rpc("GetRevenueByStore", ".payment.v1.GetRevenueByStoreRequest", ".payment.v1.RevenueByStoreResponse"),
			),
		},
	}
}

func customerProto() *descriptorpb.FileDescriptorProto {
	return &descriptorpb.FileDescriptorProto{
		Name:    pstr("customer/v1/customer.proto"),
		Package: pstr("customer.v1"),
		Options: &descriptorpb.FileOptions{
			GoPackage: pstr("github.com/enkaigaku/dvd-rental/gen/proto/customer/v1;customerv1"),
		},
		Syntax:     pstr("proto3"),
		Dependency: []string{"google/protobuf/empty.proto", "google/protobuf/timestamp.proto"},
		MessageType: []*descriptorpb.DescriptorProto{
			msg("Customer",
				int32Field("customer_id", 1),
				int32Field("store_id", 2),
				stringField("first_name", 3),
				stringField("last_name", 4),
				stringField("email", 5),
				int32Field("address_id", 6),
				boolField("active", 7),
				messageField("create_date", 8, ".google.protobuf.Timestamp"),
				messageField("last_update", 9, ".google.protobuf.Timestamp"),
				stringField("password_hash", 10),
			),
			msg("CustomerDetail",
				messageField("customer", 1, ".customer.v1.Customer"),
				stringField("address", 2),
				stringField("address2", 3),
				stringField("district", 4),
				stringField("city", 5),
				stringField("country", 6),
				stringField("postal_code", 7),
				stringField("phone", 8),
			),
			msg("GetCustomerRequest",
				int32Field("customer_id", 1),
			),
			msg("GetCustomerByEmailRequest",
				stringField("email", 1),
			),
			msg("ListCustomersRequest",
				int32Field("page_size", 1),
				int32Field("page", 2),
			),
			msg("ListCustomersResponse",
				repeatedMessageField("customers", 1, ".customer.v1.Customer"),
				int32Field("total_count", 2),
			),
			msg("ListCustomersByStoreRequest",
				int32Field("store_id", 1),
				int32Field("page_size", 2),
				int32Field("page", 3),
			),
			msg("CreateCustomerRequest",
				int32Field("store_id", 1),
				stringField("first_name", 2),
				stringField("last_name", 3),
				stringField("email", 4),
				int32Field("address_id", 5),
				boolField("active", 6),
			),
			msg("UpdateCustomerRequest",
				int32Field("customer_id", 1),
				int32Field("store_id", 2),
				stringField("first_name", 3),
				stringField("last_name", 4),
				stringField("email", 5),
				int32Field("address_id", 6),
				boolField("active", 7),
			),
			msg("DeleteCustomerRequest",
				int32Field("customer_id", 1),
			),
			msg("GetCustomerStandingRequest",
				int32Field("customer_id", 1),
			),
			msg("CustomerStanding",
				int32Field("customer_id", 1),
				boolField("in_good_standing", 2),
				repeatedStringField("reasons", 3),
				int32Field("active_rentals", 4),
				int32Field("overdue_rentals", 5),
				stringField("outstanding_balance", 6),
			),
			msg("GetCustomerSummaryRequest",
				int32Field("customer_id", 1),
			),
			msg("CustomerSummary",
				int32Field("customer_id", 1),
				int32Field("total_rentals", 2),
				int32Field("active_rentals", 3),
				stringField("total_spent", 4),
				stringField("favorite_category", 5),
				stringField("outstanding_balance", 6),
			),
			msg("Address",
				int32Field("address_id", 1),
				stringField("address", 2),
				stringField("address2", 3),
				stringField("district", 4),
				int32Field("city_id", 5),
				stringField("postal_code", 6),
				stringField("phone", 7),
				messageField("last_update", 8, ".google.protobuf.Timestamp"),
			),
			msg("GetAddressRequest",
				int32Field("address_id", 1),
			),
			msg("ListAddressesRequest",
				int32Field("page_size", 1),
				int32Field("page", 2),
			),
			msg("ListAddressesResponse",
				repeatedMessageField("addresses", 1, ".customer.v1.Address"),
				int32Field("total_count", 2),
			),
			msg("CreateAddressRequest",
				stringField("address", 1),
				stringField("address2", 2),
				stringField("district", 3),
				int32Field("city_id", 4),
				stringField("postal_code", 5),
				stringField("phone", 6),
			),
			msg("UpdateAddressRequest",
				int32Field("address_id", 1),
				stringField("address", 2),
				stringField("address2", 3),
				stringField("district", 4),
				int32Field("city_id", 5),
				stringField("postal_code", 6),
				stringField("phone", 7),
			),
			msg("DeleteAddressRequest",
				int32Field("address_id", 1),
			),
			msg("City",
				int32Field("city_id", 1),
				stringField("city", 2),
				int32Field("country_id", 3),
				messageField("last_update", 4, ".google.protobuf.Timestamp"),
			),
			msg("GetCityRequest",
				int32Field("city_id", 1),
			),
			msg("ListCitiesRequest",
				int32Field("page_size", 1),
				int32Field("page", 2),
			),
			msg("ListCitiesResponse",
				repeatedMessageField("cities", 1, ".customer.v1.City"),
				int32Field("total_count", 2),
			),
			msg("Country",
				int32Field("country_id", 1),
				stringField("country", 2),
				messageField("last_update", 3, ".google.protobuf.Timestamp"),
			),
			msg("GetCountryRequest",
				int32Field("country_id", 1),
			),
			msg("ListCountriesRequest",
				int32Field("page_size", 1),
				int32Field("page", 2),
			),
			msg("ListCountriesResponse",
				repeatedMessageField("countries", 1, ".customer.v1.Country"),
				int32Field("total_count", 2),
			),
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			svc("CustomerService",
				rpc("GetCustomer", ".customer.v1.GetCustomerRequest", ".customer.v1.CustomerDetail"),
				rpc("GetCustomerByEmail", ".customer.v1.GetCustomerByEmailRequest", ".customer.v1.Customer"),
				rpc("ListCustomers", ".customer.v1.ListCustomersRequest", ".customer.v1.ListCustomersResponse"),
				rpc("ListCustomersByStore", ".customer.v1.ListCustomersByStoreRequest", ".customer.v1.ListCustomersResponse"),
				rpc("CreateCustomer", ".customer.v1.CreateCustomerRequest", ".customer.v1.Customer"),
				rpc("UpdateCustomer", ".customer.v1.UpdateCustomerRequest", ".customer.v1.Customer"),
				rpc("DeleteCustomer", ".customer.v1.DeleteCustomerRequest", ".google.protobuf.Empty"),
				rpc("GetCustomerStanding", ".customer.v1.GetCustomerStandingRequest", ".customer.v1.CustomerStanding"),
				rpc("GetCustomerSummary", ".customer.v1.GetCustomerSummaryRequest", ".customer.v1.CustomerSummary"),
			),
			svc("AddressService",
				rpc("GetAddress", ".customer.v1.GetAddressRequest", ".customer.v1.Address"),
				rpc("ListAddresses", ".customer.v1.ListAddressesRequest", ".customer.v1.ListAddressesResponse"),
				rpc("CreateAddress", ".customer.v1.CreateAddressRequest", ".customer.v1.Address"),
				rpc("UpdateAddress", ".customer.v1.UpdateAddressRequest", ".customer.v1.Address"),
				rpc("DeleteAddress", ".customer.v1.DeleteAddressRequest", ".google.protobuf.Empty"),
			),
			svc("CityService",
				rpc("GetCity", ".customer.v1.GetCityRequest", ".customer.v1.City"),
				rpc("ListCities", ".customer.v1.ListCitiesRequest", ".customer.v1.ListCitiesResponse"),
			),
			svc("CountryService",
				rpc("GetCountry", ".customer.v1.GetCountryRequest", ".customer.v1.Country"),
				rpc("ListCountries", ".customer.v1.ListCountriesRequest", ".customer.v1.ListCountriesResponse"),
			),
		},
	}
}

func rentalProto() *descriptorpb.FileDescriptorProto {
	return &descriptorpb.FileDescriptorProto{
		Name:    pstr("rental/v1/rental.proto"),
		Package: pstr("rental.v1"),
		Options: &descriptorpb.FileOptions{
			GoPackage: pstr("github.com/enkaigaku/dvd-rental/gen/proto/rental/v1;rentalv1"),
		},
		Syntax:     pstr("proto3"),
		Dependency: []string{"google/protobuf/empty.proto", "google/protobuf/timestamp.proto"},
		MessageType: []*descriptorpb.DescriptorProto{
			msg("Rental",
				int32Field("rental_id", 1),
				messageField("rental_date", 2, ".google.protobuf.Timestamp"),
				int32Field("inventory_id", 3),
				int32Field("customer_id", 4),
				messageField("return_date", 5, ".google.protobuf.Timestamp"),
				int32Field("staff_id", 6),
				messageField("last_update", 7, ".google.protobuf.Timestamp"),
			),
			msg("RentalDetail",
				messageField("rental", 1, ".rental.v1.Rental"),
				stringField("customer_name", 2),
				stringField("film_title", 3),
				int32Field("store_id", 4),
			),
			msg("GetRentalRequest",
				int32Field("rental_id", 1),
			),
			msg("ListRentalsRequest",
				int32Field("page_size", 1),
				int32Field("page", 2),
			),
			msg("ListRentalsResponse",
				repeatedMessageField("rentals", 1, ".rental.v1.Rental"),
				int32Field("total_count", 2),
			),
			msg("ListRentalsByCustomerRequest",
				int32Field("customer_id", 1),
				int32Field("page_size", 2),
				int32Field("page", 3),
			),
			msg("ListRentalsByInventoryRequest",
				int32Field("inventory_id", 1),
				int32Field("page_size", 2),
				int32Field("page", 3),
			),
			msg("ListOverdueRentalsRequest",
				int32Field("page_size", 1),
				int32Field("page", 2),
			),
			msg("CreateRentalRequest",
				int32Field("inventory_id", 1),
				int32Field("customer_id", 2),
				int32Field("staff_id", 3),
			),
			msg("ReturnRentalRequest",
				int32Field("rental_id", 1),
			),
			msg("ReturnRentalResponse",
				messageField("rental", 1, ".rental.v1.Rental"),
				stringField("late_fee", 2),
				int32Field("days_overdue", 3),
			),
			msg("DeleteRentalRequest",
				int32Field("rental_id", 1),
			),
			msg("Inventory",
				int32Field("inventory_id", 1),
				int32Field("film_id", 2),
				int32Field("store_id", 3),
				messageField("last_update", 4, ".google.protobuf.Timestamp"),
			),
			msg("GetInventoryRequest",
				int32Field("inventory_id", 1),
			),
			msg("ListInventoryRequest",
				int32Field("page_size", 1),
				int32Field("page", 2),
			),
			msg("ListInventoryResponse",
				repeatedMessageField("items", 1, ".rental.v1.Inventory"),
				int32Field("total_count", 2),
			),
			msg("ListInventoryByFilmRequest",
				int32Field("film_id", 1),
				int32Field("page_size", 2),
				int32Field("page", 3),
			),
			msg("ListInventoryByStoreRequest",
				int32Field("store_id", 1),
				int32Field("page_size", 2),
				int32Field("page", 3),
			),
			msg("CheckInventoryAvailabilityRequest",
				int32Field("inventory_id", 1),
			),
			msg("CheckInventoryAvailabilityResponse",
				boolField("available", 1),
			),
			msg("ListAvailableInventoryRequest",
				int32Field("film_id", 1),
				int32Field("store_id", 2),
				int32Field("page_size", 3),
				int32Field("page", 4),
			),
			msg("CreateInventoryRequest",
				int32Field("film_id", 1),
				int32Field("store_id", 2),
			),
			msg("DeleteInventoryRequest",
				int32Field("inventory_id", 1),
			),
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			svc("RentalService",
				rpc("GetRental", ".rental.v1.GetRentalRequest", ".rental.v1.RentalDetail"),
				rpc("ListRentals", ".rental.v1.ListRentalsRequest", ".rental.v1.ListRentalsResponse"),
				rpc("ListRentalsByCustomer", ".rental.v1.ListRentalsByCustomerRequest", ".rental.v1.ListRentalsResponse"),
				rpc("ListRentalsByInventory", ".rental.v1.ListRentalsByInventoryRequest", ".rental.v1.ListRentalsResponse"),
				rpc("ListOverdueRentals", ".rental.v1.ListOverdueRentalsRequest", ".rental.v1.ListRentalsResponse"),
				rpc("CreateRental", ".rental.v1.CreateRentalRequest", ".rental.v1.Rental"),
				rpc("ReturnRental", ".rental.v1.ReturnRentalRequest", ".rental.v1.ReturnRentalResponse"),
				rpc("DeleteRental", ".rental.v1.DeleteRentalRequest", ".google.protobuf.Empty"),
			),
			svc("InventoryService",
				rpc("GetInventory", ".rental.v1.GetInventoryRequest", ".rental.v1.Inventory"),
				rpc("ListInventory", ".rental.v1.ListInventoryRequest", ".rental.v1.ListInventoryResponse"),
				rpc("ListInventoryByFilm", ".rental.v1.ListInventoryByFilmRequest", ".rental.v1.ListInventoryResponse"),
				rpc("ListInventoryByStore", ".rental.v1.ListInventoryByStoreRequest", ".rental.v1.ListInventoryResponse"),
				rpc("CheckInventoryAvailability", ".rental.v1.CheckInventoryAvailabilityRequest", ".rental.v1.CheckInventoryAvailabilityResponse"),
				rpc("ListAvailableInventory", ".rental.v1.ListAvailableInventoryRequest", ".rental.v1.ListInventoryResponse"),
				rpc("CreateInventory", ".rental.v1.CreateInventoryRequest", ".rental.v1.Inventory"),
				rpc("DeleteInventory", ".rental.v1.DeleteInventoryRequest", ".google.protobuf.Empty"),
			),
		},
	}
}

func main() {
	protos := map[string]*descriptorpb.FileDescriptorProto{
		"payment":  paymentProto(),
		"customer": customerProto(),
		"rental":   rentalProto(),
	}
	for name, fdp := range protos {
		encoded := encodeRawDesc(fdp)
		fmt.Fprintf(os.Stdout, "=== %s rawDesc ===\n%s\n\n", name, encoded)
	}
}
