package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	res, err := store.Get(id)
	require.NoError(t, err)

	assert.Equal(t, res.Address, parcel.Address)
	assert.Equal(t, res.Status, parcel.Status)
	assert.Equal(t, res.CreatedAt, parcel.CreatedAt)
	assert.Equal(t, res.Client, parcel.Client)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Equal(t, sql.ErrNoRows, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)

	defer db.Close()

	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	res, err := db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES ( :client, :status, :address, :createdAt)",
		sql.Named("client", parcel.Client),
		sql.Named("status", parcel.Status),
		sql.Named("address", parcel.Address),
		sql.Named("createdAt", parcel.CreatedAt))
	require.NoError(t, err)

	id, err := res.LastInsertId()
	require.NoError(t, err)
	require.NotZero(t, id)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	_, err = db.Exec("UPDATE parcel SET address = :address WHERE number = :number", sql.Named("address", newAddress), sql.Named("number", parcel.Number))
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	row := db.QueryRow("SELECT address FROM parcel WHERE number = :number", sql.Named("number", parcel.Number))

	var getAddress string
	err = row.Scan(&getAddress)
	require.NoError(t, err)
	require.Equal(t, newAddress, getAddress)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)

	defer db.Close()

	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	res, err := db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :createdAt)",
		sql.Named("client", parcel.Client),
		sql.Named("status", parcel.Status),
		sql.Named("address", parcel.Address),
		sql.Named("createdAt", parcel.CreatedAt))
	require.NoError(t, err)

	id, err := res.LastInsertId()
	require.NoError(t, err)
	require.NotZero(t, id)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := ParcelStatusSent
	_, err = db.Exec("UPDATE parcel SET status = :status WHERE number = :number", sql.Named("status", newStatus), sql.Named("number", parcel.Number))
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	row := db.QueryRow("SELECT status FROM parcel WHERE number = :number", sql.Named("number", parcel.Number))

	var getStatus string
	err = row.Scan(&getStatus)
	require.NoError(t, err)
	require.Equal(t, getStatus, newStatus)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)

	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.NotZero(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err)
	require.Equal(t, len(storedParcels), len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		value, ok := parcelMap[parcel.Number]
		require.True(t, ok)

		require.Equal(t, value.Client, parcel.Client)
		require.Equal(t, value.Status, parcel.Status)
		require.Equal(t, value.Address, parcel.Address)
		require.Equal(t, value.CreatedAt, parcel.CreatedAt)

	}
}
